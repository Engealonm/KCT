# KCT - KazCryptoToken

Welcome to the **KazCryptoToken**, an innovative blockchain market platform designed to ensure your financial freedom and privacy.

## Project Overview

KCT aims to provide users with a secure and efficient platform for blockchain transactions and asset management. Our platform focuses on:
- Ensuring user privacy.
- Facilitating seamless cryptocurrency operations.
- Empowering individuals with financial independence.

## Team Members
- [Zekenov Meiirzhan]
- [Akimzhanov Almansur]


## Screenshot of the Homepage
![image](https://github.com/user-attachments/assets/f29c1fb1-d085-411d-bd9b-8360d324661d)


## How to Run the Project

### Step 1: Clone the Repository
  Open a terminal or command prompt.
  
  Use the following command to clone the repository:
  git clone <repository-url>
  Replace <repository-url> with the URL of your Git repository.
  
  Navigate to the project directory:
  cd <repository-name>
### Step 2: Set Up the Backend
  Install Dependencies
  Ensure you have Go and MongoDB installed. If not, follow these steps:
  Download and install Go.
  Install MongoDB.
  Install Go dependencies by running:
  go mod tidy
  Start MongoDB
  Ensure MongoDB is running on localhost:27017. Use:
  sudo service mongod start   # For Linux
  brew services start mongodb # For macOS
  If using Docker:
  Run the Backend Server
  Start the Go server:
  go run main.go
  If successful, you’ll see:
  Connected to MongoDB!
  Server running on port 8080
### Step 3: Access the Frontend
  Open a web browser.
  Navigate to:
  http://localhost:8080
  This serves your static files from the static folder.



## Tools and Resources Used
- **Frontend**: HTML, JS, CSS
- **Backend**: Golang
- **Database**: MongoDB
- **Testing**: Postman



