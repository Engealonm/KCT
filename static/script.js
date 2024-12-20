document.querySelector(".buy a").addEventListener("click", function (event) {
    event.preventDefault();
    fetch("/api", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ message: "Buy crypto request" }),
    })
        .then((response) => response.json())
        .then((data) => {
            console.log(data);
            alert(data.message);
        })
        .catch((error) => {
            console.error("Error:", error);
        });
});
