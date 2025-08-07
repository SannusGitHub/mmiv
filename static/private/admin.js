/*
    fetch('/api/pinPost', {
        method: "POST",
        header: new Headers({
            "Content-Type": "application/json",
        }),
        body: JSON.stringify({
            id: postId,
            pinned: isPinned
        })
    }).then(response => {
        if (!response.ok) {
            throw new Error("Failed");
        }
        return response.json();
    }).then(data => {
        console.log("Success:", data);

        requestPosts();
    }).catch(error => {
        console.error("Error:", error);
    });
*/
