function fetchFunction(
    apiUrl, 
    method = 'POST',
    headers = { 'Content-Type': 'application/json' },
    data = {},
    responseFunc = (response) => response.json(),
    runFunc = (data) => {},
    catchFunc = (error) => {}
) {
    fetch(apiUrl, {
        method: method,
        headers: headers,
        body: method !== 'GET' ? JSON.stringify(data) : undefined
    })
    .then(response => {
        if (!response.ok) throw new Error("Network response was not ok");
        return responseFunc(response);
    })
    .then(data => runFunc(data))
    .catch(error => catchFunc(error));
};

document.addEventListener("DOMContentLoaded", function () {
    const form = document.getElementById("login-form");

    form.addEventListener("submit", function (e) {
        e.preventDefault();
        const username = document.getElementById("username").value;
        const password = document.getElementById("password").value;

        fetchFunction(
            "/api/login", 
            'POST', 
            { 'Content-Type': 'application/json' }, 
            { 
                username: username, 
                password: password
            },
            (response) => response.json(),
            (data) => {
                window.location.href = "/";
            },
            (error) => {
                console.error("Login failed:", error);
                alert("Login failed: Invalid username or password.");
            }
        );
    });
});

/*
fetchFunction(
    "/api/addUser", 
    'POST', 
    { 'Content-Type': 'application/json' }, 
    { 
        username: "sannu", 
        password: "admin",
        rank: "2",
    }
)
    */