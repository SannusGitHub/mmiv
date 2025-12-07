function returnButton() {
    document.getElementById('return-button').addEventListener('click', function() {
        window.location.href = "/";
    });
};

/*
    TODO: 
    * fix bug where overwrite on back-end doesn't work for some reason
    * clean up UI
*/
async function announcementHandler() {
    const announcementForm = document.getElementById('announcement-form');

    announcementForm.onsubmit = (e) => {
        e.preventDefault();

        const announcementContent = document.getElementById('announcement-content');
        const formData = new FormData(announcementForm);
        formData.append("content", announcementContent.value);

        fetch('/api/addAnnouncement', {
            method: "POST",
            body: formData,
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    errorText.textContent = text || response.statusText;
                    throw new Error(text || response.statusText);
                });
            }
            return response.json();
        })
        .then(data => console.log("Success:", data))
        .catch(error => console.error("Error:", error));
    };

    document.getElementById('remove-announcement-button').addEventListener('click', function() {
        fetch('/api/removeAnnouncement', {
            method: 'GET',
        }).then(res => {
            if (!res.ok) {
                throw new Error("Failed");
            }
            return res.json();
        }).then(data => {
            console.log("Success:", data);
        }).catch(error => {
            console.error("Error:", error);
        });
    });
};

document.addEventListener("DOMContentLoaded", (event) => {
    returnButton();
    announcementHandler();
});