const API_URL = 'http://localhost:8080';

// Отображение токенов на странице
function renderTokens(tokens) {
    console.log('Rendering tokens:', tokens); // Для отладки
    const tokenList = document.getElementById('tokens');
    tokenList.innerHTML = '';
    tokens.forEach(token => {
        const li = document.createElement('li');
        li.textContent = `ID: ${token.id} | Name: ${token.name} | Amount: ${token.amount}`;
        const deleteBtn = document.createElement('button');
        deleteBtn.style.color = "black";
        deleteBtn.textContent = 'Delete';
        deleteBtn.onclick = () => deleteToken(token.id);
        li.appendChild(deleteBtn);
        tokenList.appendChild(li);
    });
}

// Функция получения токенов с сортировкой
async function fetchSortedTokens(order) {
    try {
        const response = await fetch(`${API_URL}/tokens?sortOrder=${order}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch tokens: ${response.statusText}`);
        }
        const data = await response.json();

        if (data.status === 'success') {
            renderTokens(data.data);
        } else {
            alert(data.message);
        }
    } catch (error) {
        console.error('Error fetching sorted tokens:', error);
    }
}


// Обработчики событий для кнопок сортировки
document.getElementById('sortAsc').addEventListener('click', () => {
    console.log('Sort Ascending clicked'); // Для отладки
    fetchSortedTokens('asc'); // Сортировка по возрастанию
});

document.getElementById('sortDesc').addEventListener('click', () => {
    console.log('Sort Descending clicked'); // Для отладки
    fetchSortedTokens('desc'); // Сортировка по убыванию
});

// Первоначальная загрузка токенов без сортировки
async function fetchTokens() {
    try {
        const response = await fetch(`${API_URL}/tokens`);
        if (!response.ok) {
            throw new Error(`Failed to fetch tokens: ${response.statusText}`);
        }
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
console.log('Rendering tokens:', tokens);

// Вызываем при загрузке страницы
fetchTokens();
