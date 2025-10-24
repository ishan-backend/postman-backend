package comments

import (
	"errors"
	"sync"
)

// service struct
type CommentService struct{}

type ICommentService interface {
	// parentCommentId = "" if comment is made directly on post
	// else parentCommentId will be supplied from request body
	AddCommentToPost(userId string, commentString string, parentCommentId string, postId string) (bool, error)
	GetAllCommentsForPost(postId string) (AllComments, error)
}

// type response model
type AllComments struct {
	Depth    int // 
	CommentStrings []CommentResponse
	Comments []AllComments
}

type CommentResponse struct {
	CommentString string
	CommentID string
}

/*
 - 0 <AllComments, CommentStrings>
   -
    - 1
	  - 2
	    - 3
		-
		-
		-
		-
	  -
	  -
	  -
	-
	-
	-
   -
   -
   -


*/

// data models
type Post struct {
	Id      string
	Content string
}

type User struct {
	Id       string
	UserName string
}

type Comment struct {
	Id            string
	CommentString string
	ParentComment string
	Replies       *[]Comment
	PostId        string
}

// in-memory data structure
type InMemoryStore struct {
	mu               sync.RWMutex
	UserInfo         map[string]*User
	PostInfo         map[string]*Post
	PostCommentsInfo map[string][]*Comment
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		UserInfo:         make(map[string]*User),
		PostInfo:         make(map[string]*Post),
		PostCommentsInfo: make(map[string][]*Comment),
	}
}

func (i *InMemoryStore) AddCommentToPost(userId string, commentString string, parentCommentId string, postId string) (bool, error) {
	// i.mu.RWLock()
	// defer i.mu.Unlock()

	if _, ok := i.UserInfo[userId]; !ok {
		return false, errors.New("user does not exist!")
	}

	if _, ok := i.PostInfo[postId]; !ok {
		return false, errors.New("post does not exist!")
	}

	// first search for this parentCommentId in replies of this post
	// once that is found add this to datastore
	if commentString == "" {
		// add new comment to post
		comment := make([]Comment, 0)
		i.PostCommentsInfo[postId] = append(i.PostCommentsInfo[postId], &Comment{
			Id:            "", // random UUID
			CommentString: commentString,
			ParentComment: "",
			Replies:       &comment,
			PostId:        postId,
		})

		return true, nil
	}

	// add nested comment to post
	if val, ok := i.PostCommentsInfo[postId]; ok {
		commentAdded := false
		for _, parentComment := range val {
			if parentComment != nil {
				commentAdded = commentAdded && dfs(*parentComment, commentString, parentCommentId)
			}
		}

		if commentAdded {
			return true, nil
		}
	}

	return false, errors.New("parent commentId not found")

}

func dfs(comment Comment, commentString string, parentCommentId string) bool {
	if comment.Id == parentCommentId {
		// we have found parent comment, add this here now
		newReplies := make([]Comment, 0)
		originalCOmments := *comment.Replies
		originalCOmments = append(originalCOmments, Comment{
			Id:            "", // new id,
			CommentString: commentString,
			ParentComment: comment.Id,
			Replies:       &newReplies,
		})

		comment.Replies = &originalCOmments
		return true
	}

	// dfs for all children comments
	addedComment := false
	for _, reply := range *comment.Replies {
		addedComment = addedComment && dfs(reply, commentString, parentCommentId)
	}
	return addedComment
}

func (i *InMemoryStore) GetAllCommentsForPost(postId string) (AllComments, error) {
	
	if _, ok : =i.PostCommentsInfo[postId]; !ok {
		return AllComments{}, errors.New("postId is not available in records")
	}

	ans := AllComments{Depth: 0, CommentStrings: make([]string, 0)}
	for _, comment := range i.PostCommentsInfo[postId] {
		ans.Comments = append(ans.Comments, dfsAllComments(0, comment))
	}

	return ans
}

func dfsAllComments(depth int, comment Comment) AllComments {
	newDepth := depth+1
	finalCommentStrings := make([]string, 0)
	finalComments := make([]AllComments, 0)

	for _, reply := range *comment.Replies {
		finalCommentStrings = append(finalCommentStrings, reply.CommentString)
		finalComments = append(finalComments, dfsAllComments(newDepth, reply)...)
	}

	ans := AllComments{}
	ans.Depth = newDepth
	ans.CommentString = finalCommentStrings
	ans.Comments = finalComments

	return ans
}
