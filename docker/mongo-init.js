// Initialize single-node replica set for enabling transactions
rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "localhost:27017" }
  ]
});
// Wait until PRIMARY
var retries = 0;
while (retries < 30) {
  var status = rs.status();
  if (status.ok === 1 && status.members && status.members[0].stateStr === "PRIMARY") {
    break;
  }
  sleep(1000);
  retries++;
}

