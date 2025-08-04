/*
    LIST OF THINGS TO CURRENTLY DO / BUGS PRESENT:
    * rewrite the comments into supporting new multi-part form system & images (X)
        front-end (X) and back-end (X)
    * fix autorequesting new posts whenever a post / comment is submitted (X)
    * general clean-up (man what the fuck is this mess)
*/

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

/*
fetchFunction(
    "/api/addUser", 
    'POST', 
    { 'Content-Type': 'application/json' }, 
    { 
        username: "newUser", 
        password: "testpassword",
        rank: "1",
    }
)
*/

function logout() {
    fetchFunction(
        "/api/logout", 
        'POST', 
        { 'Content-Type': 'application/json' }, 
        {},
        (response) => response.json(),
        (data) => {
            window.location.href = "/login";
        },
        (error) => {
            console.error("Logout failed:", error);
            alert("Logout failed.");
        }
    );
};

function requestPosts() {
    setupDraggableForm({
        grabBarLabelText: 'New Post',
        formButtonLabelText: 'Post',
        apiEndpoint: '/api/addPost',
        getPayload: () => ({
            postcontent: document.getElementById("post-content").value
        }),
        onSuccess: () => requestPosts()
    });

    fetchFunction(
        "/api/requestPost", 
        'POST', 
        { 'Content-Type': 'application/json' },
        {},
        (response) => response.json(),
        (data) => {
            const postsContainer = document.getElementById('content');
            postsContainer.innerHTML = "";

            data.forEach((element) => {
                createPost(
                    element,
                    function() {
                        requestComments(element);
                    }
                );
            });
        },
        (error) => {
            console.error("Error fetching post:", error);
        }
    );
};

function requestComments(parentpost) {
    setupDraggableForm({
        grabBarLabelText: 'Add Comment',
        formButtonLabelText: 'Comment',
        apiEndpoint: '/api/addComment',
        getPayload: () => ({
            parentpostid: parentpost.id,
            postcontent: document.getElementById("post-content").value
        }),
        onSuccess: () => requestComments(parentpost)
    });

    const postsContainer = document.getElementById('content');
    postsContainer.innerHTML = "";

    createPost(parentpost);
    // console.log(parentpostid);

    fetchFunction(
        "/api/requestComment", 
        'POST', 
        { 'Content-Type': 'application/json' },
        {
            parentpostid: parentpost.id,
        },
        (response) => response.json(),
        (data) => {
            data.forEach((element) => {
                createPost(element);
            });
        },
        (error) => {
            console.error("Error fetching post:", error);
        }
    );
};

function setupDraggableForm({
    grabBarLabelText = null,
    formButtonLabelText = null,
    apiEndpoint,
    getPayload,
    onSuccess = () => {},
}) {
    const form = document.getElementById("post-form");
    const grabBar = document.getElementById("grabBar");
    const fileInput = document.getElementById("post-image");
    const textInput = document.getElementById("post-content");

    if (grabBarLabelText !== null) {
        grabBar.textContent = grabBarLabelText;
    }

    let offsetX = 0, offsetY = 0, isDragging = false;

    grabBar.onmousedown = (e) => {
        isDragging = true;
        const rect = form.getBoundingClientRect();
        offsetX = e.clientX - rect.left;
        offsetY = e.clientY - rect.top;
        document.body.style.userSelect = 'none';
    };

    document.onmousemove = (e) => {
        if (isDragging) {
            form.style.left = `${e.clientX - offsetX}px`;
            form.style.top = `${e.clientY - offsetY}px`;
        }
    };

    document.onmouseup = () => {
        isDragging = false;
        document.body.style.userSelect = '';
    };

    // handle post button
    const postButton = document.getElementById("post-button");
    postButton.innerText = formButtonLabelText;
    postButton.onmousedown = (e) => {
        if (form.style.display === "none") {
            form.style.display = "block";
        } else {
            form.style.display = "none";
        }
    };

    form.onsubmit = (e) => {
        e.preventDefault();
        
        const formData = new FormData();
        formData.append("postcontent", textInput.value);
        formData.append("image", fileInput.files[0]);

        fetch(apiEndpoint, {
            method: "POST",
            body: formData,
        }).then(response => {
            if (!response.ok) {
                throw new Error("Upload failed");
            }
            return response.json();
        }).then(data => {
            console.log("Success:", data);
        }).catch(error => {
            console.error("Error:", error);
        });
    };
}

/*
COMMENT FETCH

fetchFunction(
            "/api/addPost", 
            'POST', 
            { 'Content-Type': 'application/json' }, 
            {
                postcontent: postcontent
            },
            (response) => response.json(),
            (data) => {
                document.getElementById("comment-content").value = "";
                // requestComments();
            }
        );

*/

// id, timestamp, user, content
function createPost(data, clickFunc) {
    const postDiv = document.createElement('div');
    postDiv.className = 'accented';
    postDiv.setAttribute('postId', data.id);

    const titleP = document.createElement('p');
    titleP.className = 'header';
    titleP.innerHTML = `#${data.id} <span class="highlight"><b>${data.username}</b></span> @ ${data.timestamp}`;

    const postContentDiv = document.createElement('div');
    postContentDiv.className = "post-content";

    console.log(data.imagepath);
    if (data.imagepath !== null && data.imagepath !== "") {
        const postImg = document.createElement('img');
        postImg.src = data.imagepath;
        postImg.className = 'image-content';
        postContentDiv.appendChild(postImg);
    };

    const contentP = document.createElement('p');
    contentP.className = 'text-content';
    contentP.textContent = data.postcontent;
    postContentDiv.appendChild(contentP);

    postDiv.appendChild(titleP);
    postDiv.appendChild(postContentDiv);

    const postsContainer = document.getElementById('content');
    postsContainer.appendChild(postDiv);

    if (typeof clickFunc === 'function') {
        postDiv.classList.add("clickable");
        postDiv.addEventListener('click', function() {
            clickFunc();
        });
    }
};

document.addEventListener("DOMContentLoaded", function() { 
    requestPosts();
});