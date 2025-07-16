const loginForm = document.getElementById("loginForm");

loginForm.addEventListener("submit", async function (event) {
    event.preventDefault();
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    try {
        const response = await fetch(`/submit/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            email: email,
            password: password,
        }),
        });

        if (!response.ok) {
            alert("Invalid credentials!");
            return;
        }
        alert("Login sucessfully!");
        const data = await response.json();
        const jwtToken = data.token;

        localStorage.setItem("jwtToken", jwtToken);
        // window.location.href = '/';
    }
    catch (error) {
        console.error('Error during login:', error);
        alert('An error occurred. Please try again.');
    }
});
