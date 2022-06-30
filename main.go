package main

import (
	"context"
	// "database/sql"
	"encoding/json"
	"net/http"
	"os"

	// "strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	// "github.com/xuri/excelize/v2"
)

var ctx = context.Background()

const expCookieTime = 1382400

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

// var psqlconn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
// 	"185.225.35.60", 5432, "postgres", "1991055q", "glasha")

type Auth struct {
	Email    string
	Password string
}

// func importXLSX() {
// 	f, err := excelize.OpenFile("./bd.xlsm")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	db, _ := sql.Open("postgres", psqlconn)

// 	comp_id_cols, _ := f.GetCols("Сопоставление ID")
// 	// budgets_cols, _ := f.GetCols("Budgets")
// 	// payments_students_cols, _ := f.GetCols("Payments_Students")

// 	// comp_id_students_index := 0
// 	comp_id_budgets_index := 1
// 	comp_id_logins_index := 4
// 	comp_id_passwords_index := 5

// 	// budgets_name_index = 0
// 	// budgets_budgets_index = 1

// 	insert_users := `INSERT INTO "user" ("login", "password") VALUES `

// 	n := 0
// 	for i, val := range comp_id_cols[comp_id_logins_index] {
// 		if val != "" {
// 			insert_users += `('` + comp_id_cols[comp_id_logins_index][i] + `','` +
// 				comp_id_cols[comp_id_passwords_index][i] + `'),`
// 			// comp_id_cols[comp_id_logins_index][n] = val
// 			// comp_id_cols[comp_id_passwords_index][n] = comp_id_cols[comp_id_passwords_index][i]
// 			comp_id_cols[comp_id_budgets_index][n] = comp_id_cols[comp_id_budgets_index][i]
// 			n++
// 		}
// 	}
// 	comp_id_cols[comp_id_logins_index] = comp_id_cols[comp_id_logins_index][1:n]
// 	comp_id_cols[comp_id_passwords_index] = comp_id_cols[comp_id_passwords_index][1:n]
// 	comp_id_cols[comp_id_budgets_index] = comp_id_cols[comp_id_budgets_index][1:n]

// 	db.Exec(`DELETE FROM "user";`)

// 	_, err = db.Exec(strings.TrimSuffix(insert_users, ",") + `;`)
// 	fmt.Print(err)
// 	// for _, col := range cols {
// 	// 	for _, rowCell := range col {
// 	// 		fmt.Print(rowCell, "\t")
// 	// 	}
// 	// }
// 	defer db.Close()
// 	f.Close()
// }

func createUserSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://185.225.35.60")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var a Auth
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionID, _ := uuid.NewRandom()

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  sessionID.String(),
		MaxAge: expCookieTime,
	}

	_, err = rdb.SetNX(ctx, sessionID.String(), 0, expCookieTime*time.Second).Result()
	if err != nil {
		w.WriteHeader(523)
		json.NewEncoder(w).Encode(err)
		return
	}

	http.SetCookie(w, cookie)
}

func deleteUserSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://185.225.35.60")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	tokenCookie, err := r.Cookie("session-id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rdb.Del(ctx, tokenCookie.Value).Val()

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, cookie)
}

func checkUserSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://185.225.35.60")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	tokenCookie, err := r.Cookie("session-id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	val := rdb.Exists(ctx, tokenCookie.Value).Val()
	if val < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rdb.Expire(ctx, tokenCookie.Value, expCookieTime*time.Second)

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  tokenCookie.Value,
		MaxAge: expCookieTime,
	}

	http.SetCookie(w, cookie)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	mux := http.NewServeMux()

	// importXLSX()

	mux.HandleFunc("/login", createUserSession)
	mux.HandleFunc("/logout", deleteUserSession)
	mux.HandleFunc("/validate_user", checkUserSession)
	http.ListenAndServe(":"+port, mux)
}
