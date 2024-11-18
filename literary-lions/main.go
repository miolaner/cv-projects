package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob(filepath.Join("templates/*.html")))

// Helper function to generate a Universally Unique Identifiers (UUID) session ID
func generateSessionID() string {
	return uuid.New().String() // Generates a UUID, e.g., "6f1c4e86-593f-47c3-bf96-44b1fe746bab"
}

// Helper function to get user ID from session
func getUserIDFromSession(r *http.Request) int {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return 0 // Not logged in
    }

    var userID int
    err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", cookie.Value).Scan(&userID)
    if err != nil {
        return 0 // Invalid session
    }

    return userID
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Printf("Template rendering error: %v", err)
		// Do not call http.Error here; handle errors where you call renderTemplate
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)
	var username string
	var loggedIn bool

	if userID != 0 {
		err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		loggedIn = true
	} else {
		loggedIn = false
	}

	rows, err := db.Query(`
        SELECT p.id, u.username, p.category, p.title 
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
        LIMIT 5`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []struct {
		ID       int
		Username string
		Category string
		Title    string
	}
	for rows.Next() {
		var post struct {
			ID       int
			Username string
			Category string
			Title    string
		}
		err := rows.Scan(&post.ID, &post.Username, &post.Category, &post.Title)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Add My Posts and Liked Posts links only for logged in users
	var categories []struct {
		Value string
		Name  string
	}

	categories = []struct {
		Value string
		Name  string
	}{
		{"", "All Categories"},
		{"Reviews", "Reviews"},
		{"Questions", "Questions"},
		{"Marketplace", "Marketplace"},
		{"Random", "Random"},
	}

	if loggedIn {
		categories = append(categories, struct {
			Value string
			Name  string
		}{"my-posts", "My Posts"})

		categories = append(categories, struct {
			Value string
			Name  string
		}{"liked-posts", "Liked Posts"})
	}

	data := struct {
		Username   string
		LoggedIn   bool
		Categories []struct {
			Value string
			Name  string
		}
		Posts []struct {
			ID       int
			Username string
			Category string
			Title    string
		}
	}{
		Username:   username,
		LoggedIn:   loggedIn,
		Categories: categories,
		Posts:      posts,
	}

	renderTemplate(w, "home", data)
}

// Handler for user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Example structure to pass data to the template
		data := struct {
			ErrorMessage string
			Email        string
			Username     string
		}{
			Email:    email,
			Username: username,
		}

		// Validate email format
		if !isValidEmail(email) {
			data.ErrorMessage = "Invalid email format."
			renderTemplate(w, "register", data)
			return
		}

		// Check if email is already in use
		var existingUserID int
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&existingUserID)
		if err == nil {
			data.ErrorMessage = "Email is already registered."
			renderTemplate(w, "register", data)
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if username is already in use
		err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&existingUserID)
		if err == nil {
			data.ErrorMessage = "Username is already taken."
			renderTemplate(w, "register", data)
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Insert the new user into the database
		_, err = db.Exec("INSERT INTO users (email, username, password, is_moderator) VALUES (?, ?, ?, FALSE)", email, username, string(hashedPassword))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		renderTemplate(w, "register", nil)
	}
}

// Handler for user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var hashedPassword string
		var userID int

		// First, check if the username exists
		err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
		if err == sql.ErrNoRows {
			// Username does not exist
			data := struct {
				Message string
			}{
				Message: "Invalid username",
			}
			renderTemplate(w, "login", data)
			return
		} else if err != nil {
			// Some internal error occurred
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Next, verify the password
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			// Password is incorrect
			data := struct {
				Message string
			}{
				Message: "Invalid password",
			}
			renderTemplate(w, "login", data)
			return
		}
		// Generate session ID and log the user in
		sessionID := generateSessionID()
		expiresAt := time.Now().Add(24 * time.Hour)
		_, err = db.Exec("INSERT INTO sessions (user_id, session_id, expires_at) VALUES (?, ?, ?)", userID, sessionID, expiresAt)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
            Name:    "session_id",
            Value:   sessionID,
            Expires: expiresAt,
            Path:    "/",
            HttpOnly: true,
        })

		// Check for the redirect URL
		redirectURL := r.URL.Query().Get("redirect")
		content := r.URL.Query().Get("content")

		if redirectURL != "" {
			if content != "" {
				// Re-attach the comment content to the redirect URL
				redirectURL += "&content=" + url.QueryEscape(content)
			}
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	} else {
		// Check for the 'message' query parameter
		message := r.URL.Query().Get("message")
		redirectURL := r.URL.Query().Get("redirect")

		// Prepare the message based on the value of the query parameter
		var errorMsg string
		switch message {
		case "login_required_comment":
			errorMsg = "Please log in first before commenting."
		case "login_required_like":
			errorMsg = "Please log in first before liking a post."
		case "login_required_create_post":
			errorMsg = "Please log in first before creating a post."
		default:
			errorMsg = ""
		}

		// Pass the message to the template
		data := struct {
			Message     string
			RedirectURL string
		}{
			Message:     errorMsg,
			RedirectURL: redirectURL,
		}

		// Render the login template with the message
		renderTemplate(w, "login", data)
	}
}

// Handler for user logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = db.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
        Name:    "session_id",
        Value:   "",
        Expires: time.Now().Add(-1 * time.Hour),
        Path:    "/",
    })

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// Handler to create a new post
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)
	if userID == 0 {
		// Redirect to login page with a message and redirect to create post after login
		http.Redirect(w, r, "/login?message=login_required_create_post&redirect=/create-post", http.StatusSeeOther)
		return
	}
	if r.Method == "POST" {
		category := r.FormValue("category")
		title := r.FormValue("title")
		content := r.FormValue("content")

		if category != "Reviews" && category != "Questions" && category != "Marketplace" && category != "Random" {
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}

		// Insert the post into the database and get the ID of the newly created post
		result, err := db.Exec("INSERT INTO posts (user_id, category, title, content) VALUES (?, ?, ?, ?)", userID, category, title, content)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Get the ID of the newly inserted post
		postID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to the newly created post
		http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10), http.StatusSeeOther)
	} else {
		renderTemplate(w, "create_post", nil)
	}
}

// Handler to create a new comment
func createCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)
	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	if userID == 0 {
		// Redirect to login page with a message and the post URL
		http.Redirect(w, r, "/login?message=login_required_comment&redirect=/post?id="+postID+"&content="+url.QueryEscape(content), http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		// Check if the comment content is empty
		if content == "" {
			// Redirect to post page with an error message
			http.Redirect(w, r, "/post?id="+postID+"&error=empty_comment", http.StatusSeeOther)
			return
		}

		// Insert the comment into the database
		_, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
	}
}

// Handler to like or dislike a post/comment
func likeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userID := getUserIDFromSession(r)
		if userID == 0 {
			// Redirect to login page with a message and the post URL
			postID := r.FormValue("post_id")
			http.Redirect(w, r, "/login?message=login_required_like&redirect=/post?id="+postID, http.StatusSeeOther)
			return
		}

		postID := r.FormValue("post_id")
		like := r.FormValue("like") == "true"

		// Check if the user has already liked or disliked this post
		var existingLike bool
		err := db.QueryRow("SELECT like FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLike)

		if err == sql.ErrNoRows {
			// User hasn't liked/disliked yet, insert the new like/dislike
			_, err := db.Exec("INSERT INTO likes (user_id, post_id, like) VALUES (?, ?, ?)", userID, postID, like)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else if err == nil {
			// User has already liked/disliked, update their existing like/dislike
			_, err := db.Exec("UPDATE likes SET like = ? WHERE user_id = ? AND post_id = ?", like, userID, postID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect back to the post page
		http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	var title, content, category, username string
	var createdAt time.Time
	var creatorID int

	err := db.QueryRow(`
        SELECT p.title, p.content, p.category, p.created_at, u.username, u.id 
        FROM posts p 
        JOIN users u ON p.user_id = u.id 
        WHERE p.id = ?`, postID).Scan(&title, &content, &category, &createdAt, &username, &creatorID)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Query for like and dislike counts
	var likeCount, dislikeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND like = true", postID).Scan(&likeCount)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND like = false", postID).Scan(&dislikeCount)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the current user ID from session
	userID := getUserIDFromSession(r)

	// Check if the current user has liked or disliked the post
	var userLikeStatus *bool
	err = db.QueryRow("SELECT like FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&userLikeStatus)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch and render existing comments
	rows, err := db.Query(`
    SELECT c.id, c.content, c.created_at, u.username 
    FROM comments c
    JOIN users u ON c.user_id = u.id
    WHERE c.post_id = ?
    ORDER BY c.created_at ASC`, postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []struct {
		ID        string
		Content   string
		CreatedAt string
		Username  string
	}
	for rows.Next() {
		var comment struct {
			ID        string
			Content   string
			CreatedAt time.Time
			Username  string
		}
		err := rows.Scan(&comment.ID, &comment.Content, &comment.CreatedAt, &comment.Username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		comments = append(comments, struct {
			ID        string
			Content   string
			CreatedAt string
			Username  string
		}{
			ID:        comment.ID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format("Jan 2, 2006 at 3:04pm"),
			Username:  comment.Username,
		})
	}

	// Check if there is an error for an empty comment
	var errorMsg string
	if r.URL.Query().Get("error") == "empty_comment" {
		errorMsg = "Comment cannot be empty."
	}

	// Check if there's a pre-filled comment content (after login redirect)
	prefillComment := r.URL.Query().Get("content")

	// Prepare the data for the template
	data := struct {
		ID             string
		Title          string
		Category       string
		Content        string
		CreatedAt      string
		LikeCount      int
		DislikeCount   int
		UserLikeStatus struct {
			Liked    bool
			Disliked bool
		}
		Error    string
		Comments []struct {
			ID        string
			Content   string
			CreatedAt string
			Username  string
		}
		PrefillComment  string // Field to pre-fill the comment
		CreatorID       int
		CreatorUsername string
	}{
		ID:              postID,
		Title:           title,
		Category:        category,
		Content:         content,
		CreatedAt:       createdAt.Format("Jan 2, 2006 at 3:04pm"),
		LikeCount:       likeCount,
		DislikeCount:    dislikeCount,
		Error:           errorMsg,
		PrefillComment:  prefillComment, // Add the pre-filled comment
		CreatorID:       creatorID,
		CreatorUsername: username,
	}

	if userLikeStatus != nil {
		data.UserLikeStatus.Liked = *userLikeStatus
		data.UserLikeStatus.Disliked = !*userLikeStatus
	}

	data.Comments = comments

	var isModerator bool
	db.QueryRow("SELECT is_moderator FROM users WHERE id = ?", userID).Scan(&isModerator)

	// Render the appropriate template based on moderator status
	if isModerator {
		renderTemplate(w, "post", data)
	} else {
		renderTemplate(w, "post_guest", data)
	}
}


func filterPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)
	var username string
	var loggedIn bool
	filter := r.FormValue("filter")
	var title string
	var rows *sql.Rows
	var err error

	if userID != 0 {
		err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		loggedIn = true
	} else {
		loggedIn = false
	}
// Determine the query based on the filter
// Determine the query based on the filter
if filter == "my-posts" && userID != 0 {
	rows, err = db.Query(`
		SELECT p.id, u.username, p.category, p.title 
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		WHERE p.user_id = ?`, userID)
	title = "My Posts"
} else if filter == "liked-posts" && userID != 0 {
	rows, err = db.Query(`
		SELECT p.id, u.username, p.category, p.title 
		FROM posts p
		JOIN users u ON p.user_id = u.id 
		JOIN likes l ON p.id = l.post_id 
		WHERE l.user_id = ? AND l.like = true`, userID)
	title = "Liked Posts"
} else if filter != "" {
	rows, err = db.Query(`
		SELECT p.id, u.username, p.category, p.title 
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		WHERE p.category = ?`, filter)
	title = filter // You can map this to more user-friendly titles if necessary
} else {
	rows, err = db.Query(`
		SELECT p.id, u.username, p.category, p.title 
		FROM posts p 
		JOIN users u ON p.user_id = u.id`)
	title = "All Posts"
}

if err != nil {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
}
defer rows.Close()

	var posts []struct {
		ID       int
		Username string
		Category string
		Title    string
	}
	for rows.Next() {
		var post struct {
			ID       int
			Username string
			Category string
			Title    string
		}
		err := rows.Scan(&post.ID, &post.Username, &post.Category, &post.Title)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Add My Posts and Liked Posts links only for logged in users
	var categories []struct {
		Value string
		Name  string
	}

	categories = []struct {
		Value string
		Name  string
	}{
		{"", "All Categories"},
		{"Reviews", "Reviews"},
		{"Questions", "Questions"},
		{"Marketplace", "Marketplace"},
		{"Random", "Random"},
	}

	if loggedIn {
		categories = append(categories, struct {
			Value string
			Name  string
		}{"my-posts", "My Posts"})

		categories = append(categories, struct {
			Value string
			Name  string
		}{"liked-posts", "Liked Posts"})
	}

	data := struct {
		Title string
		Username   string
		LoggedIn   bool
		Categories []struct {
			Value string
			Name  string
		}
		Posts []struct {
			ID       int
			Username string
			Category string
			Title    string
		}
	}{
		Title: title,
		Username:   username,
		LoggedIn:   loggedIn,
		Categories: categories,
		Posts:      posts,
	}
	renderTemplate(w, "filter_posts", data)
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	field := r.URL.Query().Get("field")
	if field == "" {
		field = "all" // Default to searching all fields if no specific field is chosen
	}

	data := struct {
		Query        string
		Field        string
		Posts        []struct {
			ID       int
			Username string
			Category string
			Title    string
		}
		ErrorMessage string
	}{
		Query: query,
		Field: field,
	}

	if query == "" {
		data.ErrorMessage = "Can't find any results for empty search."
		renderTemplate(w, "search_results", data)
		return
	}

	var rows *sql.Rows
	var err error

	// Adjust the query based on the selected field
	switch field {
	case "title":
		rows, err = db.Query(`
			SELECT p.id, u.username, p.category, p.title
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.title LIKE ?
			ORDER BY p.created_at DESC`, "%"+query+"%")
	case "content":
		rows, err = db.Query(`
			SELECT p.id, u.username, p.category, p.title
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.content LIKE ?
			ORDER BY p.created_at DESC`, "%"+query+"%")
	case "creator":
		rows, err = db.Query(`
			SELECT p.id, u.username, p.category, p.title
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE u.username LIKE ?
			ORDER BY p.created_at DESC`, "%"+query+"%")
	case "all":
		rows, err = db.Query(`
			SELECT p.id, u.username, p.category, p.title
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.title LIKE ? OR p.content LIKE ? OR u.username LIKE ?
			ORDER BY p.created_at DESC`, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	default:
		data.ErrorMessage = "Invalid search field selected."
		renderTemplate(w, "search_results", data)
		return
	}

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post struct {
			ID       int
			Username string
			Category string
			Title    string
		}
		err := rows.Scan(&post.ID, &post.Username, &post.Category, &post.Title)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.Posts = append(data.Posts, post)
	}

	renderTemplate(w, "search_results", data)
}



func profileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromSession(r)
	var profileID int
	var email sql.NullString
	var firstName, lastName, aboutMe sql.NullString
	var posts []struct {
		ID    int
		Title string
	}

	// Fetch user information
	err := db.QueryRow(`
        SELECT id, email, first_name, last_name, about_me 
        FROM users 
        WHERE username = ?
    `, username).Scan(&profileID, &email, &firstName, &lastName, &aboutMe)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Println("Error: User not found")
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error fetching user data:", err)
		return
	}

	// Fetch the 5 most recent posts of the user
	rows, err := db.Query("SELECT id, title FROM posts WHERE user_id = ? ORDER BY created_at DESC LIMIT 5", profileID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error fetching user posts:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post struct {
			ID    int
			Title string
		}
		err := rows.Scan(&post.ID, &post.Title)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error scanning post:", err)
			return
		}
		posts = append(posts, post)
	}

	// Convert NULL values to empty strings
	firstNameStr := ""
	if firstName.Valid {
		firstNameStr = firstName.String
	}

	lastNameStr := ""
	if lastName.Valid {
		lastNameStr = lastName.String
	}

	aboutMeStr := ""
	if aboutMe.Valid {
		aboutMeStr = aboutMe.String
	}

	data := struct {
		Username  string
		FirstName string
		LastName  string
		Email     string
		AboutMe   string
		Posts     []struct {
			ID    int
			Title string
		}
		IsOwner bool
	}{
		Username:  username,
		FirstName: firstNameStr,
		LastName:  lastNameStr,
		Email:     email.String,
		AboutMe:   aboutMeStr,
		Posts:     posts,
		IsOwner:   userID == profileID,
	}

	renderTemplate(w, "profile", data)

}

func editProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var username string
	var firstName, lastName, aboutMe sql.NullString
	var errorMsg string

	// Fetch the username at the beginning
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error fetching username:", err)
		return
	}

	if r.Method == "POST" {
		firstName = sql.NullString{String: r.FormValue("first_name"), Valid: true}
		lastName = sql.NullString{String: r.FormValue("last_name"), Valid: true}
		aboutMe = sql.NullString{String: r.FormValue("about_me"), Valid: true}

		// Handle password change
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if newPassword != "" {
			// Verify the current password
			var storedPasswordHash string
			err := db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&storedPasswordHash)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Println("Error fetching current password hash:", err)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(currentPassword))
			if err != nil {
				errorMsg = "Current password is incorrect"
			} else if newPassword != confirmPassword {
				errorMsg = "New passwords do not match"
			} else {
				// Hash the new password
				newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Println("Error hashing new password:", err)
					return
				}

				// Update the password in the database
				_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", newHashedPassword, userID)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Println("Error updating password:", err)
					return
				}
			}
		}

		// Update profile information if no password error
		if errorMsg == "" {
			_, err := db.Exec("UPDATE users SET first_name = ?, last_name = ?, about_me = ? WHERE id = ?", firstName, lastName, aboutMe, userID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Println("Error updating user profile:", err)
				return
			}

			// Redirect to the updated profile page
			http.Redirect(w, r, "/profile?username="+username, http.StatusSeeOther)
			return
		}
	} else {
		err := db.QueryRow("SELECT first_name, last_name, about_me FROM users WHERE id = ?", userID).Scan(&firstName, &lastName, &aboutMe)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error fetching user profile:", err)
			return
		}
	}

	// Convert NULL values to empty strings
	firstNameStr := ""
	if firstName.Valid {
		firstNameStr = firstName.String
	}

	lastNameStr := ""
	if lastName.Valid {
		lastNameStr = lastName.String
	}

	aboutMeStr := ""
	if aboutMe.Valid {
		aboutMeStr = aboutMe.String
	}

	data := struct {
		Username  string
		FirstName string
		LastName  string
		AboutMe   string
		ErrorMsg  string
	}{
		Username:  username,
		FirstName: firstNameStr,
		LastName:  lastNameStr,
		AboutMe:   aboutMeStr,
		ErrorMsg:  errorMsg,
	}

	renderTemplate(w, "edit_profile", data)
}
func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Check if the user is a moderator
	userID := getUserIDFromSession(r)
	var isModerator bool
	db.QueryRow("SELECT is_moderator FROM users WHERE id = ?", userID).Scan(&isModerator)

	if !isModerator {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	commentID := r.FormValue("comment_id")
	if commentID == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page with success status
	http.Redirect(w, r, "/post?id="+r.FormValue("post_id"), http.StatusSeeOther)
}
func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Check if the user is a moderator
	userID := getUserIDFromSession(r)
	var isModerator bool
	db.QueryRow("SELECT is_moderator FROM users WHERE id = ?", userID).Scan(&isModerator)

	if !isModerator {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page with success status
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	initDB()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/create-post", createPostHandler)
	http.HandleFunc("/create-comment", createCommentHandler)
	http.HandleFunc("/like", likeHandler)
	http.HandleFunc("/filter-posts", filterPostsHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/edit-profile", editProfileHandler)
	http.HandleFunc("/delete-comment", deleteCommentHandler)
	http.HandleFunc("/delete-post", deletePostHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func isValidEmail(email string) bool {
	// Simple regex for email validation
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
