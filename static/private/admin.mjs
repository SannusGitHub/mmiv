import * as index from '/static/home/index.mjs';

let optionsMenu;

document.addEventListener("DOMContentLoaded", (e) => {
    optionsMenu = document.getElementById("option-menu");

    document.getElementById('pin-post-button').addEventListener('click', function(e) {
        const postId = optionsMenu.getAttribute("attached-to-id");
        pinPost(postId);
    });
});

function pinPost(postId) {
    const currentPost = index.currentPosts.get(postId);

    fetch('/api/pinPost', {
        method: "POST",
        header: new Headers({
            "Content-Type": "application/json",
        }),
        body: JSON.stringify({
            id: postId,
            pinned: !currentPost.pinned
        })
    }).then(response => {
        if (!response.ok) {
            throw new Error("Failed");
        };

        return response.json();
    }).then(data => {
        console.log("Success:", data);

        optionsMenu.style.display = "none";
        index.fetchPosts();
    }).catch(error => {
        console.error("Error:", error);
    });
}