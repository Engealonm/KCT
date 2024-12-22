const API_URL = 'http://localhost:8080';

// Fetch tokens from the server
async function fetchTokens() {
    try {
        const response = await fetch(`${API_URL}/tokens`);
        const data = await response.json();
        if (data.status === 'success') {
            renderTokens(data.data);
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error fetching tokens:', error);
    }
}

// Render tokens in the list
function renderTokens(tokens) {
    const tokenList = document.getElementById('tokens');
    tokenList.innerHTML = '';
    tokens.forEach(token => {
        const li = document.createElement('li');
        li.textContent = `${token.name} - ${token.amount}`;
        const deleteBtn = document.createElement('button');
        deleteBtn.textContent = 'Delete';
        deleteBtn.onclick = () => deleteToken(token.id);
        li.appendChild(deleteBtn);
        tokenList.appendChild(li);
    });
}

// Create a new token
document.getElementById('tokenForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('tokenName').value;
    const amount = document.getElementById('tokenAmount').value;

    try {
        const response = await fetch(`${API_URL}/tokens/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ name, amount: parseInt(amount) })
        });
        const data = await response.json();
        if (data.status === 'success') {
            alert('Token created successfully!');
            fetchTokens();
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error creating token:', error);
    }
});

// Delete a token
async function deleteToken(id) {
    try {
        const response = await fetch(`${API_URL}/tokens/delete?id=${id}`, {
            method: 'DELETE'
        });
        const data = await response.json();
        if (data.status === 'success') {
            alert('Token deleted successfully!');
            fetchTokens();
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error deleting token:', error);
    }
}
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
// Initial fetch of tokens
fetchTokens();