// Prevent default form submission for debugging
document.getElementById("loginForm").addEventListener("submit", function (event) {
    event.preventDefault(); // Prevent default form submission
    
    // Collect form data
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    // Send login request to the server
    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
    })
    .then(response => {
        // Check if the status code indicates success
        if (response.status === 200) {
            return response.json(); // Parse the JSON response
        } else if (response.status === 401) {
            throw new Error("Invalid credentials"); // Handle incorrect login
        } else {
            throw new Error(`Unexpected error: ${response.status}`);
        }
    })
    .then(data => {
        // If the response indicates success, alert the user
        if (data.success === "true") {
            alert("Login success!");
        }
    })
    .catch(error => {
        // Handle errors and display appropriate messages
        alert(error.message);
    });
});
