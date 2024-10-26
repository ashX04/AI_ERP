
├── cmd/
│ └── main.go # Application entry point and server setup
├── internal/
│ ├── handlers/ # HTTP request handlers
│ │ ├── auth.go # Authentication handlers
│ │ ├── dashboard.go # Dashboard view handlers
│ │ ├── files.go # File management handlers
│ │ └── home.go # Home page handlers
│ ├── middleware/ # Custom middleware
│ │ ├── auth.go # Authentication middleware
│ │ └── logger.go # Request logging middleware
│ ├── models/ # Data models and database interactions
│ │ ├── user.go # User model
│ │ └── file.go # File model
│ └── templates/ # HTML templates
│ ├── dashboard.html # Dashboard view
│ ├── login.html # Login form
│ ├── register.html # Registration form
│ └── upload.html # File upload form
├── static/ # Static assets
│ ├── css/ # Stylesheets
│ └── js/ # JavaScript files
└── uploads/ # Protected file storage

## 🔧 Setup & Installation

1. **Clone the Repository**
   ```bash
   git clone https://github.com/yourusername/project-name.git
   cd project-name
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Environment Setup**
   - Create necessary directories:
     ```bash
     mkdir -p uploads
     ```
   - Set up environment variables (if needed)

4. **Run the Application**
   ```bash
   go run cmd/main.go
   ```
   The server will start at `http://localhost:8080`

## 🔐 Security Features

- Session-based Authentication
- Secure Cookie Storage
- HTTP-only Cookies
- Request Logging
- Protected Routes
- File Type Validation

## 🛣️ API Routes

### Public Routes
- `GET /` - Home page
- `GET /login` - Login page
- `GET /register` - Registration page
- `POST /register` - Register new user
- `POST /login` - User login
- `GET /logout` - User logout

### Protected Routes (Requires Authentication)
- `GET /dashboard` - User dashboard
- `GET /upload` - Upload page
- `POST /upload` - Handle file upload
- `GET /download/:id` - Download file
- `DELETE /files/:id` - Delete file
- `GET /preview/:id` - Preview image

## 💻 Development

### Prerequisites
- Go 1.16 or higher
- Git

### Development Server
For hot reloading during development, you can use Air:

bash
air

### Building for Production
bash
go build -o app cmd/main.go

## 🧪 Testing

Run the test suite:
bash
go test ./...


## 📝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Support

For support, email your-email@example.com or open an issue in the repository.

## ✨ Acknowledgments

- Gin Web Framework
- HTMX
- Go community