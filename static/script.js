const API_URL = 'http://localhost:8080';

let currentPage = 1;
let totalPages = 1;
let currentSortOrder = '';
let currentMinAmount = '';
let currentMaxAmount = '';

async function fetchTokens(page = 1, limit = 5, sortOrder = currentSortOrder, minAmount = currentMinAmount, maxAmount = currentMaxAmount) {
    try {
        const queryParams = new URLSearchParams({ page, limit, sortOrder, minAmount, maxAmount }).toString();
        const response = await fetch(`${API_URL}/tokens?${queryParams}`);
        const data = await response.json();

        if (data.status === 'success') {
            renderTokens(data.data.tokens);
            currentPage = data.data.currentPage;
            totalPages = data.data.totalPages;
            renderPagination(currentPage, totalPages);
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error fetching tokens:', error);
    }
}

function renderTokens(tokens) {
    const tokenList = document.getElementById('tokens');
    tokenList.innerHTML = '';
    if (Array.isArray(tokens)) {
        tokens.forEach(token => {
            const li = document.createElement('li');
            li.textContent = `ID: ${token.id} | Name: ${token.name} | Amount: ${token.amount}`;
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = 'Delete';
            deleteBtn.onclick = () => deleteToken(token.id);
            li.appendChild(deleteBtn);
            tokenList.appendChild(li);
        });
    } else {
        alert("No tokens to display.");
    }
}

async function createToken(name, amount) {
    try {
        const response = await fetch(`${API_URL}/tokens/create`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
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
}

async function deleteToken(id) {
    try {
        const response = await fetch(`${API_URL}/tokens/delete?id=${id}`, { method: 'DELETE' });
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

async function updateToken(id, name, amount) {
    try {
        const response = await fetch(`${API_URL}/tokens/update`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id, name, amount: parseInt(amount) })
        });
        const data = await response.json();
        if (data.status === 'success') {
            alert('Token updated successfully!');
            fetchTokens();
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error updating token:', error);
    }
}

async function searchToken(id) {
    try {
        const response = await fetch(`${API_URL}/tokens/search?id=${id}`);
        const data = await response.json();
        if (data.status === 'success') {
            renderSearchedToken(data.data);
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error searching for token:', error);
    }
}

function renderSearchedToken(token) {
    const searchResult = document.getElementById('searchResult');
    searchResult.innerHTML = '';
    if (token) {
        const li = document.createElement('li');
        li.textContent = `${token.name} - ${token.amount}`;
        searchResult.appendChild(li);
    } else {
        searchResult.textContent = 'Token not found.';
    }
}

function renderPagination(currentPage, totalPages) {
    const pageInfo = document.getElementById('pageInfo');
    pageInfo.textContent = `Page ${currentPage} of ${totalPages}`;

    const prevButton = document.getElementById('prevPage');
    const nextButton = document.getElementById('nextPage');

    prevButton.disabled = currentPage === 1;
    nextButton.disabled = currentPage === totalPages;
}

document.getElementById('tokenForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('tokenName').value;
    const amount = document.getElementById('tokenAmount').value;
    await createToken(name, amount);
});

document.getElementById('searchForm').addEventListener('submit', (e) => {
    e.preventDefault();
    const tokenId = document.getElementById('searchTokenId').value;
    searchToken(tokenId);
});

document.getElementById('prevPage').addEventListener('click', () => {
    if (currentPage > 1) {
        fetchTokens(currentPage - 1);
    }
});

document.getElementById('nextPage').addEventListener('click', () => {
    if (currentPage < totalPages) {
        fetchTokens(currentPage + 1);
    }
});

document.getElementById('sortAsc').addEventListener('click', () => {
    currentSortOrder = 'asc';
    fetchTokens(1, 5, currentSortOrder);
});

document.getElementById('sortDesc').addEventListener('click', () => {
    currentSortOrder = 'desc';
    fetchTokens(1, 5, currentSortOrder);
});

document.getElementById('updateTokenForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('updateTokenId').value;
    const name = document.getElementById('updateTokenName').value;
    const amount = parseInt(document.getElementById('updateTokenAmount').value);
    await updateToken(id, name, amount);
});

document.getElementById('filterForm').addEventListener('submit', (e) => {
    e.preventDefault();
    currentMinAmount = document.getElementById('minAmount').value;
    currentMaxAmount = document.getElementById('maxAmount').value;
    currentSortOrder = document.querySelector('input[name="sortOrder"]:checked')?.value || '';
    fetchTokens(1, 5, currentSortOrder, currentMinAmount, currentMaxAmount);
});

document.getElementById('emailForm').addEventListener('submit', async function (event) {
    event.preventDefault(); // Остановить стандартное поведение формы

    // Получение данных из формы
    const to = document.getElementById('to').value;
    const subject = document.getElementById('subject').value;
    const body = document.getElementById('body').value;

    try {
        // Отправка POST-запроса на сервер
        const response = await fetch('/sendmail', { // Замените '/send-email' на ваш реальный путь
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ to, subject, body }),
        });

        if (!response.ok) {
            const errorMessage = await response.text();
            alert(`Ошибка: ${errorMessage}`);
            return;
        }

        const result = await response.json();
        if (result.success) {
            alert('Email успешно отправлен!');
            document.getElementById('emailForm').reset(); // Очистить форму
        }
    } catch (error) {
        console.error('Ошибка при отправке запроса:', error);
        alert('Ошибка при отправке email. Проверьте консоль для деталей.');
    }
});

fetchTokens();