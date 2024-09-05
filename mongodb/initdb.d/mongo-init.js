db.createUser({
    user: "root",
    pwd: "root123",
    roles: [
        {
        role: "readWrite",
        db: "mongodb",
        },
    ],
});