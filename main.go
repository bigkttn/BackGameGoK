package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
)

// กำหนดโครงสร้างข้อมูล user (ตรงกับ table user ของคุณ)
type User struct {
	UID      string `json:"uid"` // เปลี่ยนจาก int เป็น string
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"` // ถ้าตารางไม่มี Phone ให้ใช้ placeholder หรือเลือกคอลัมน์อื่น
	Role     string `json:"role"`
}

var db *sql.DB

func main() {
	// Connection string
	dsn := "66011212075:0934308887@tcp(202.28.34.210:3309)/db66011212075"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot ping database:", err)
	}
	fmt.Println("✅ Connected to database successfully")

	// สร้าง API endpoint
	http.HandleFunc("/user", getUsers)
	// เพิ่ม handler สำหรับ register
	http.HandleFunc("/register", registerUser)

	// หา IP ของเครื่อง
	ip := getLocalIP()
	url := fmt.Sprintf("http://%s:8080/user", ip)

	// เปิด browser อัตโนมัติ
	openBrowser(url)

	// run server
	fmt.Printf("🚀 Server started at %s\n", url)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// handler ดึงข้อมูล user ทั้งหมด
func getUsers(w http.ResponseWriter, r *http.Request) {
	// เลือกเฉพาะคอลัมน์ที่ตรงกับ struct
	rows, err := db.Query("SELECT uid, username AS full_name, email, '' AS phone, role FROM user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.UID, &u.FullName, &u.Email, &u.Phone, &u.Role); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// หา IPv4 LAN จริง
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				if ip[0] == 192 || ip[0] == 10 || (ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) {
					return ip.String()
				}
			}
		}
	}
	return "localhost"
}

// เปิด browser อัตโนมัติ
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin": // MacOS
		cmd = "open"
		args = []string{url}
	default: // Linux
		cmd = "xdg-open"
		args = []string{url}
	}

	exec.Command(cmd, args...).Start()
}

// handler ลงทะเบียนผู้ใช้ใหม่
func registerUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// รับข้อมูล JSON จาก body
	var u struct {
		UID      string `json:"uid"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// INSERT ลงฐานข้อมูล
	stmt, err := db.Prepare("INSERT INTO user (uid, username, email, password, role) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.UID, u.FullName, u.Email, u.Password, u.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}
