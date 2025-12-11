Forum — Go Web Application

This project is a lightweight forum built using Go, SQLite, HTML templates and Docker.
It provides user authentication, post creation, commenting, category filtering, and like/dislike functionality.
The application follows a clean handler-based architecture and uses a persistent SQLite database.

Features
Authentication

User registration

User login

User logout

Secure session management stored in SQLite

Password hashing using bcrypt

Posts

Create new posts

View an individual post

Posts display author, timestamp, categories and like/dislike counts

Logged-in users can like or dislike posts

Posts can be assigned to one or more categories

Posts are publicly viewable without logging in

Comments

Add comments to a post (authentication required)

Comments linked to the appropriate user

Like/dislike functionality for comments

Categories

Default categories are seeded automatically

Category list available on the homepage

View all posts belonging to a selected category

Likes System

A user may only like or dislike a post once

Like/dislike support for both posts and comments

Counts updated and displayed in real time

User Filters

“My Posts” — displays only posts created by the logged-in user

“Liked Posts” — displays only posts the user has liked

Error Pages

Custom 404 Not Found page

Custom 500 Internal Server Error page with panic recovery

Project Structure
/forum
│── main.go
│── go.mod
│── go.sum
│── forum.db
│
├── database/
│   ├── db.go
│   └── schema.sql
│
├── handlers/
│   ├── auth/
│   ├── posts/
│   ├── comments/
│   ├── likes/
│   └── categories/
│
├── models/
│
├── static/
│
└── templates/
    ├── index.html
    ├── post.html
    ├── create_post.html
    ├── register.html
    ├── login.html
    ├── error_404.html
    └── error_500.html

Database Schema

The database includes the following tables:

users

sessions

posts

comments

categories

post_categories (many-to-many relationship)

likes (supports posts and comments)

Key schema properties:

ON DELETE CASCADE used for all relationships

likes table enforces values of 1 or -1

post_categories prevents duplicate category associations

Session table includes expiry timestamps

Routes Overview
Public Routes
Route	Description
/	Homepage — displays all posts
/post?id=X	View a single post
/category?id=X	Display posts for a given category
/register	Create a new account
/login	User login
Authenticated Routes
Route	Description
/logout	Log out
/create-post	Create a new post
/create-comment	Add a comment
/like	Like or dislike content
/my-posts	User’s own posts
/liked-posts	Posts the user has liked
Error Routes
Route	Result
Any invalid URL	Custom 404 page
Any panic	Custom 500 page
Running the Project with Docker
Build and run (standard)
docker-compose down
docker-compose up --build

Clean build (no cache)
docker-compose down
docker-compose build --no-cache
docker-compose up

Access the application

Open the browser at:

http://localhost:8080

Testing Checklist (Audit Requirements)
Authentication

Register a new user

Log in with correct credentials

Login should fail with incorrect credentials

Logout clears the session

Posts

Create a post with at least one category

Verify it appears on the homepage

Verify the link opens the correct post page

Like/dislike functionality updates correctly

Public users can view posts but cannot like/dislike

Comments

Only logged-in users can create comments

Comments display with correct username and timestamp

Comment likes/dislikes update correctly

Comments disappear if the database is reset

Categories

Homepage displays all categories

Clicking a category filters posts

Creating a post stores assigned categories

Filters

/my-posts shows only posts by the logged-in user

/liked-posts shows only posts the user liked

Error Handling

Navigating to a non-existent route (e.g. /doesnotexist) shows the custom 404 page

Triggering a controlled internal error shows the custom 500 page

Docker

Application rebuilds successfully

Database persists across restarts (unless deleted manually)

Notes

Session management is handled via cookies and SQLite.

All foreign keys have cascading delete enabled.

Default categories are seeded using INSERT OR IGNORE.

The homepage includes a category sidebar for filtering.

All logic uses the Go standard library (no external web frameworks).

License

This project was developed as part of the 01 Founders curriculum and is intended for educational use.