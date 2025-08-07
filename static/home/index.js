let currentPosts = new Map;

class Post {
    constructor({
        id, username, postcontent, imagepath, commentcount, timestamp, pinned, clickFunc
    } = {}) {
        this.id = id;
        this.username = username;
        this.postcontent = postcontent;
        this.commentcount = commentcount;
        this.imagepath = imagepath;
        this.timestamp = timestamp;
        this.pinned = pinned;
        this.clickFunc = clickFunc;
    };

    createElements() {
        const postDiv = document.createElement('div');
        postDiv.className = 'accented';
        postDiv.setAttribute('postId', this.id);

        // header

        const headerDiv = document.createElement('div');
        headerDiv.className = 'header post-label';

        const headerRightDiv = document.createElement('div');
        headerRightDiv.className = 'right';

        const settingButtonP = document.createElement('p');
        settingButtonP.innerText = "☰";
        settingButtonP.className = 'clickable';

        const headerTitleP = document.createElement('p');
        headerTitleP.innerHTML = `#${this.id} <span class="highlight"><b>${this.username}</b></span> @ ${this.timestamp}`;
        if (this.commentcount !== undefined) {
            headerTitleP.innerHTML += ` | R: ${this.commentcount}`
        };
        if (this.pinned !== undefined && this.pinned == true) {
            headerTitleP.innerHTML += ` | pinned`
        };

        const optionsMenu = document.getElementById("option-menu");
        settingButtonP.addEventListener('click', function(e) {
            e.stopPropagation();
            const rect = settingButtonP.getBoundingClientRect();

            if (optionsMenu.style.display === "none") {
                optionsMenu.style.display = "block";
            } else {
                optionsMenu.style.display = "none";
            };

            optionsMenu.style.top = `${rect.bottom + window.scrollY}px`;
            optionsMenu.style.left = `${rect.left - optionsMenu.offsetWidth + settingButtonP.offsetWidth + window.scrollX}px`;
        });

        // content

        const postContentDiv = document.createElement('div');
        postContentDiv.className = "post-content";

        const contentP = document.createElement('p');
        contentP.className = 'text-content';
        contentP.textContent = this.postcontent;

        let contentImg = null;
        if (this.imagepath !== null && this.imagepath !== "") {
            contentImg = document.createElement('img');
            contentImg.src = this.imagepath;
            contentImg.classList = 'image-content clickable';

            contentImg.addEventListener('click', function(e) {
                e.stopPropagation();
                
                if (contentImg.style.width === contentImg.naturalWidth + "px") {
                    contentImg.style.width = "150px";
                    postContentDiv.style.flexDirection = "row";
                } else {
                    contentImg.style.width = contentImg.naturalWidth + "px";
                    postContentDiv.style.flexDirection = "column";
                };
            });
        };

        // append

        postDiv.appendChild(headerDiv);
        postDiv.appendChild(postContentDiv);

        headerDiv.appendChild(headerTitleP);
        headerDiv.append(headerRightDiv);

        headerRightDiv.appendChild(settingButtonP);

        if (contentImg !== null) {
            postContentDiv.appendChild(contentImg);
        };
        postContentDiv.appendChild(contentP);

        // interact

        if (typeof this.clickFunc === 'function') {
            postDiv.classList.add("clickable");
            postDiv.addEventListener('click', () => {
                this.clickFunc();
            });
        }

        return postDiv;
    };
};

function loadPosts() {
    const content = document.getElementById('content');
    content.innerHTML = "";

    currentPosts.forEach((element) => {
        const div = element.createElements();
        content.appendChild(div);
    });
};

function fetchPosts() {
    setupDraggableForm({
        grabBarLabelText: 'New Post',
        formButtonLabelText: 'Post',
        onSubmitForm: function(e) {
            e.preventDefault();
            
            const fileInput = document.getElementById("post-image");
            const textInput = document.getElementById("post-content");

            const formData = new FormData();
            formData.append("postcontent", textInput.value);
            formData.append("image", fileInput.files[0]);

            fetch('/api/addPost', {
                method: "POST",
                body: formData,
            }).then(response => {
                if (!response.ok) {
                    throw new Error("Upload failed");
                }
                return response.json();
            }).then(data => {
                console.log("Success:", data);

                fetchPosts();
                fileInput.value = null;
            }).catch(error => {
                console.error("Error:", error);
            });
        }
    });

    fetch('/api/requestPost', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
    }).then(response => {
        if (!response.ok) {
            throw new Error("Failed");
        };
        
        return response.json();
    }).then(data => {
        console.log("Success:", data);
        
        currentPosts.clear();
        data.forEach((element) => {
            const newPost = new Post({
                id: element.id,
                username: element.username,
                postcontent: element.postcontent,
                imagepath: element.imagepath,
                commentcount: element.commentcount,
                timestamp: element.timestamp,
                pinned: element.pinned,
                clickFunc: function() {
                    fetchComments(element);
                }
            });

            currentPosts.set(newPost.id, newPost);
        });

        loadPosts();
    }).catch(error => {
        console.error("Error:", error);
    });
};

function fetchComments(postParent) {
    setupDraggableForm({
        grabBarLabelText: 'Add Comment',
        formButtonLabelText: 'Comment',
        onSubmitForm: function(e) {
            e.preventDefault();

            const fileInput = document.getElementById("post-image");
            const textInput = document.getElementById("post-content");
            const parentpostID = postParent.id;
            
            const formData = new FormData();
            formData.append("postcontent", textInput.value);
            formData.append("image", fileInput.files[0]);
            formData.append("parentpostid", parentpostID);

            fetch('/api/addComment', {
                method: "POST",
                body: formData,
            }).then(response => {
                if (!response.ok) {
                    throw new Error("Upload failed");
                }
                return response.json();
            }).then(data => {
                console.log("Success:", data);

                fetchComments(postParent);
                fileInput.value = null;
            }).catch(error => {
                console.error("Error:", error);
            });
        }
    });
    
    fetch('/api/requestComment', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            parentpostid: postParent.id
        })
    }).then(response => {
        if (!response.ok) {
            throw new Error("Failed");
        };
        
        return response.json();
    }).then(data => {
        console.log("Success:", data);
        currentPosts.clear();

        const parentPost = new Post(postParent);
        currentPosts.set(parentPost.id, parentPost);
        if (Array.isArray(data)) {
            data.forEach((element) => {
                const newPost = new Post({
                    id: element.id,
                    username: element.username,
                    postcontent: element.postcontent,
                    imagepath: element.imagepath,
                    commentcount: element.commentcount,
                    timestamp: element.timestamp,
                    pinned: element.pinned
                });

                currentPosts.set(newPost.id, newPost);
            });
        };

        loadPosts();
    }).catch(error => {
        console.error("Error:", error);
    });
};

function logout() {
    /*
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
    */

    fetch('/api/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
    }).then(response => {
        if (!response.ok) {
            throw new Error("Failed");
        };
        
        return response.json();
    }).then(data => {
        window.location.href = "/login";
    }).catch(error => {
        console.error("Error:", error);
    });
};

function setupDraggableForm({
    grabBarLabelText = null,
    formButtonLabelText = null,
    onSubmitForm
}) {
    const form = document.getElementById("post-form");
    const grabBar = document.getElementById("grabBar");
    const fileInput = document.getElementById("post-image");
    const textInput = document.getElementById("post-content");
    fileInput.value = null;
    textInput.value = "";

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
        onSubmitForm(e);
    };
};

fetchPosts();