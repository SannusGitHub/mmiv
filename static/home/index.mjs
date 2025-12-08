export let currentPosts = new Map;

class Post {
    constructor({
        id,
        parentpost,
        username,
        postcontent,
        imagepath,
        commentcount,
        timestamp,
        pinned,
        locked,
        canpin,
        canlock,
        hasownership,
        iscomment,
        clickFunc
    } = {}) {
        this.id = id;
        this.parentpost = parentpost;
        this.username = username;
        this.postcontent = postcontent;
        this.commentcount = commentcount;
        this.imagepath = imagepath;
        this.timestamp = timestamp;
        this.pinned = pinned;
        this.locked = locked;
        this.canpin = canpin;
        this.canlock = canlock;
        this.hasownership = hasownership;
        this.iscomment = iscomment,
        this.clickFunc = clickFunc;
    };

    createElements() {
        // vars

        const postId = this.id;
        const parentPost = this.parentpost;
        const isPinned = this.pinned;
        const isLocked = this.locked;
        const isComment = this.iscomment;

        // post

        const postDiv = document.createElement('div');
        postDiv.className = 'accented';

        // dropdown

        const dropdownDiv = document.createElement('div');
        dropdownDiv.className = 'accented post-submit';
        dropdownDiv.style.display = "none";
        dropdownDiv.style.width = "2em";

        let deleteOption = null;
        if (this.hasownership) {
            deleteOption = document.createElement('p');
            deleteOption.className = "clickable";
            deleteOption.innerText = "Delete";
            
            let url = '/api/deletePost';
            if (isComment) {
                url = '/api/deleteComment';
            }

            deleteOption.addEventListener('click', function(e) {
                fetch(url, {
                    method: "POST",
                    header: new Headers({
                        "Content-Type": "application/json",
                    }),
                    body: JSON.stringify({
                        id: postId,
                    })
                }).then(response => {
                    if (!response.ok) {
                        throw new Error("Failed");
                    };
            
                    return response.json();
                }).then(data => {
                    console.log("Success:", data);

                    if (isComment) {
                        fetchComments(parentPost);
                    } else {
                        fetchPosts();
                    }
                }).catch(error => {
                    console.error("Error:", error);
                });
            });
        };

        let pinOption = null;
        if (this.canpin) {
            pinOption = document.createElement('p');
            pinOption.className = "clickable";
            pinOption.innerText = "Pin";

            pinOption.addEventListener('click', function(e) {
                fetch('/api/pinPost', {
                    method: "POST",
                    header: new Headers({
                        "Content-Type": "application/json",
                    }),
                    body: JSON.stringify({
                        id: postId,
                        pinned: !isPinned
                    })
                }).then(response => {
                    if (!response.ok) {
                        throw new Error("Failed");
                    };
            
                    return response.json();
                }).then(data => {
                    console.log("Success:", data);
            
                    // optionsMenu.style.display = "none";
                    fetchPosts();
                }).catch(error => {
                    console.error("Error:", error);
                });
            });
        };

        let lockOption = null;
        if (this.canlock) {
            lockOption = document.createElement('p');
            lockOption.className = "clickable";
            lockOption.innerText = "Lock";

            lockOption.addEventListener('click', function(e) {
                fetch('/api/lockPost', {
                    method: "POST",
                    header: new Headers({
                        "Content-Type": "application/json",
                    }),
                    body: JSON.stringify({
                        id: postId,
                        locked: !isLocked
                    })
                }).then(response => {
                    if (!response.ok) {
                        throw new Error("Failed");
                    };
            
                    return response.json();
                }).then(data => {
                    console.log("Success:", data);
            
                    // optionsMenu.style.display = "none";
                    fetchPosts();
                }).catch(error => {
                    console.error("Error:", error);
                });
            });
        };

        // dropdownDiv.id = "option-menu";

        // header

        const headerDiv = document.createElement('div');
        headerDiv.className = 'header post-label';

        const headerRightDiv = document.createElement('div');
        headerRightDiv.className = 'right';

        const settingButtonP = document.createElement('p');
        settingButtonP.innerText = "☰";
        settingButtonP.className = 'clickable option-button';
        settingButtonP.setAttribute("data-post-id", this.id);

        const headerTitleP = document.createElement('p');
        const localTimestamp = new Date(this.timestamp);
        const pad = num => String(num).padStart(2, "0");

        const day = pad(localTimestamp.getDate());
        const month = pad(localTimestamp.getMonth() + 1);
        const year = String(localTimestamp.getFullYear()).slice(-2);
        const hours = pad(localTimestamp.getHours());
        const minutes = pad(localTimestamp.getMinutes());
        const seconds = pad(localTimestamp.getSeconds());
        const formattedTimestamp = `${day}/${month}/${year} ${hours}:${minutes}:${seconds}`;

        headerTitleP.innerHTML = `#${this.id} <span class="highlight"><b>${this.username}</b></span> @ ${formattedTimestamp} `;
        if (this.commentcount !== undefined) {
            headerTitleP.innerHTML += `<img src="/static/img/icons/reply.png" alt="R" class="emoticon"> ${this.commentcount} `
        };

        if (this.pinned !== undefined && this.pinned == true) {
            postDiv.setAttribute("pinned", this.pinned)
            headerTitleP.innerHTML += `<img src="/static/img/icons/sticky.png" alt="P" class="emoticon"> `
        };
        
        if (this.locked !== undefined && this.locked == true) {
            postDiv.setAttribute("locked", this.locked);
            headerTitleP.innerHTML += `<img src="/static/img/icons/lock.png" alt="L" class="emoticon"> `
        };

        settingButtonP.addEventListener('click', function(e) {
            e.stopPropagation();

            // dropdownDiv.setAttribute("attached-to-id", this.getAttribute('data-post-id'));
            const rect = settingButtonP.getBoundingClientRect();

            if (dropdownDiv.style.display === "none") {
                dropdownDiv.style.display = "block";
            } else {
                dropdownDiv.style.display = "none";
            };

            dropdownDiv.style.top = `${rect.bottom + window.scrollY}px`;
            dropdownDiv.style.left = `${rect.left - dropdownDiv.offsetWidth + settingButtonP.offsetWidth + window.scrollX}px`;
        });

        // content

        const postContentDiv = document.createElement('div');
        postContentDiv.className = "post-content";

        const contentP = document.createElement('p');
        contentP.className = 'text-content';
        contentP.innerHTML = this.postcontent;

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
                    postContentDiv.style.flexWrap = "nowrap";
                } else {
                    contentImg.style.width = contentImg.naturalWidth + "px";
                    postContentDiv.style.flexDirection = "column";
                    postContentDiv.style.flexWrap = "wrap";
                };
            });
        };

        // append

        postDiv.appendChild(headerDiv);
        postDiv.appendChild(postContentDiv);

        postDiv.appendChild(dropdownDiv);
        if (deleteOption !== null) {
            dropdownDiv.appendChild(deleteOption);
        };
        if (pinOption !== null) {
            dropdownDiv.appendChild(pinOption);
        }
        if (lockOption !== null) {
            dropdownDiv.appendChild(lockOption);
        };

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

export function fetchPosts() {
    setupDraggableForm({
        grabBarLabelText: 'New Post',
        formButtonLabelText: 'Post',
        onSubmitForm: function(e) {
            e.preventDefault();
            
            const fileInput = document.getElementById("post-image");
            const textInput = document.getElementById("post-content");
            const lockInput = document.getElementById("lock-post");
            const pinInput = document.getElementById("pin-post");
            const anonInput = document.getElementById("anonymous-post");
            const errorText = document.getElementById("error-text");

            const formData = new FormData();
            formData.append("postcontent", textInput.value);
            formData.append("image", fileInput.files[0]);
            formData.append("locked", lockInput?.checked ?? false);
            formData.append("pinned", pinInput?.checked ?? false);
            formData.append("isanonymous", anonInput?.checked ?? false);

            fetch('/api/addPost', {
                method: "POST",
                body: formData,
            }).then(response => {
                if (!response.ok) {
                    return response.text().then(text => {
                        errorText.textContent = text || response.statusText;
                        throw new Error(text || response.statusText);
                    });
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

    // DISPLAY POST NUMBER AND AMOUNT OF POSTS CODE, MOVE THIS TO
    // A SMALL UI ELEMENT LATER POTENTIALLY
    const requestFormData = new FormData();
    requestFormData.append("displayfrompostnumber", 1);
    requestFormData.append("amountofpostsrequested", 20);

    fetch('/api/requestPost', {
        method: 'POST',
        body: requestFormData
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
                locked: element.locked,
                canpin: element.canpin,
                canlock: element.canlock,
                hasownership: element.hasownership,
                iscomment: element.iscomment,
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

    /*
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
                locked: element.locked,
                canpin: element.canpin,
                canlock: element.canlock,
                hasownership: element.hasownership,
                iscomment: element.iscomment,
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
    */

    const returnButton = document.getElementById('return-button');
    returnButton.style = "display: none";
};

function fetchComments(postParent) {
    setupDraggableForm({
        grabBarLabelText: 'Add Comment',
        formButtonLabelText: 'Comment',
        onSubmitForm: function(e) {
            e.preventDefault();

            const fileInput = document.getElementById("post-image");
            const textInput = document.getElementById("post-content");
            const anonInput = document.getElementById("anonymous-post");
            const errorText = document.getElementById("error-text");
            const parentpostID = postParent.id;
            
            const formData = new FormData();
            formData.append("postcontent", textInput.value);
            formData.append("image", fileInput.files[0]);
            formData.append("parentpostid", parentpostID);
            formData.append("isanonymous", anonInput?.checked ?? false);

            fetch('/api/addComment', {
                method: "POST",
                body: formData,
            }).then(response => {
                if (!response.ok) {
                    return response.text().then(text => {
                        errorText.textContent = text || response.statusText;
                        throw new Error(text || response.statusText);
                    });
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
    
    // const optionsMenu = document.getElementById('option-menu');
    // optionsMenu.style.display = "none";

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
                    parentpost: postParent,
                    username: element.username,
                    postcontent: element.postcontent,
                    imagepath: element.imagepath,
                    commentcount: element.commentcount,
                    timestamp: element.timestamp,
                    pinned: element.pinned,
                    locked: element.locked,
                    canpin: element.canpin,
                    canlock: element.canlock,
                    hasownership: element.hasownership,
                    iscomment: element.iscomment,
                });

                currentPosts.set(newPost.id, newPost);
            });
        };

        loadPosts();
    }).catch(error => {
        console.error("Error:", error);
    });

    const returnButton = document.getElementById('return-button');
    returnButton.style = "display: block";
};

function fetchAnnouncement() {
    fetch('/api/requestAnnouncement', {
        method: 'GET',
    }).then(res => {
        if (!res.ok) {
            throw new Error("Failed");
        }
        return res.json();
    }).then(data => {
        console.log("Success:", data);
        
        const announcement = document.getElementById("announcement");
        if (data.content == "") {
            announcement.remove();
            return;
        };

        announcement.style = "display: block;";
        document.getElementById("announcement-text").innerHTML = data.content;
    }).catch(error => {
        console.error("Error:", error);
    });
}

function returnButton() {
    const returnButton = document.getElementById('return-button');
    returnButton.addEventListener('click', function() {
        // const optionsMenu = document.getElementById('option-menu');
        // optionsMenu.style.display = "none";

        fetchPosts();
    });
};

function dashboardButton() {
    const dashboardButton = document.getElementById('dashboard-button');

    if (dashboardButton) {
        dashboardButton.addEventListener('click', function() {
            window.location.href = "/dashboard";
        });
    };
};

function logoutButton() {
    document.getElementById('logout-button').addEventListener('click', function() {
        fetch('/api/logout', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
        }).then(response => {
            if (!response.ok) {
                throw new Error("Failed");
            };
            
            return response.json();
        }).then(() => {
            window.location.href = "/login";
        });
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

document.addEventListener("DOMContentLoaded", (event) => {
    returnButton();
    dashboardButton();
    logoutButton();
    fetchPosts();
    fetchAnnouncement();
});