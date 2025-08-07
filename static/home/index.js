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

fetchPosts();