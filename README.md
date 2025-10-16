# postman-backend
Postman backend

docker compose up -d
docker-compose down -v // remove all volumes
mongodb://admin:admin123@localhost:27017/?authSource=admin // connection string for mongoDB compass (download this separately)


Short answer: True MongoDB multi-document transactions only work on a replica set or sharded cluster. There isn’t a way to use them on a standalone server.
Practical options:
Enable single-node replica set (what we did): run mongod with --replSet and connect with ?replicaSet=rs0. Use sessions + StartTransaction with writeConcern: majority and readConcern: snapshot.
Use a real multi-node replica set in prod for stronger guarantees and availability.
