# 📚 Library Management System (Go + Fiber + MongoDB)

A simple and efficient **Library Management API** built using **Go (Golang)**, **Fiber framework**, and **MongoDB**.  
It supports user registration, login, book borrowing and returning, and secure password hashing with **bcrypt**.

---

## 🚀 Features

- 📖 Add, list, borrow, and return books
- 👤 Register and authenticate users
- 🔐 Passwords are securely hashed using bcrypt
- 🧩 Built with modular, clean code
- ⚡ Fast RESTful API using Fiber
- 🗃️ MongoDB for storing users and books

---

## 🛠️ Tech Stack

- **Go** – Main programming language
- **Fiber** – Web framework
- **MongoDB** – NoSQL database
- **Bcrypt** – Password hashing

---

## 📂 Project Structure

```
Library-Management/
├── main.go           # Main application logic
├── go.mod / go.sum   # Go dependencies
```

---

## 🔧 Setup Instructions

### 1️⃣ Clone the repo
```bash
git clone https://github.com/YOUR_USERNAME/Library-Management.git
cd Library-Management
```

### 2️⃣ Start MongoDB
Make sure you have MongoDB running locally on port `27017`.

### 3️⃣ Run the application
```bash
go run main.go
```

The server will run on `http://localhost:3000`.

---

## 📬 API Endpoints

| Method | Endpoint                | Description               |
|--------|-------------------------|---------------------------|
| POST   | `/register`             | Register a new user       |
| POST   | `/login`                | Login with credentials    |
| GET    | `/user/:id`             | Get user info             |
| DELETE | `/user/:id`             | Delete a user             |
| POST   | `/book`                 | Add a new book            |
| GET    | `/books`                | List all books            |
| POST   | `/borrow/:bookId`       | Borrow a book             |
| POST   | `/return/:bookId`       | Return a borrowed book    |

---

## 👨‍💻 Author

Developed by **Salih Efehan Demir**  
🔗 [GitHub](https://github.com/SalihEfehanDemir)  
🔗 [LinkedIn](https://www.linkedin.com/in/salih-efehan-demir/)

---

## 📜 License

This project is open-source and available under the MIT License.
