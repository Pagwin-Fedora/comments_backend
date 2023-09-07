CREATE TABLE comments(
    -- postgres seems to have SERIAL so we can use that instead of this being a primary key
    id INTEGER NOT NULL PRIMARY KEY,
    blog_post TEXT NOT NULL,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    email_verified INTEGER NOT NULL,
    website TEXT,
    comment TEXT NOT NULL
);
