/* Updates the "totalEntry" element's value in index.html hero */
document.addEventListener('DOMContentLoaded', function () {
    fetch('https://columbus.elmasy.com/api/stat')
        .then((response) => (
            response.json()
        ))
        .then((data) => {
            console.log(data)
            this.getElementById('totalEntry').textContent = data.total
        });
}, false);