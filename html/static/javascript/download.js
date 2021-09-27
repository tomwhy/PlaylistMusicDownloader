function download_songs(playlist_id) {
    download_page(playlist_id, "")
}

function download_page(playlist_id, page) {
    fetch(`/api/download/${playlist_id}/${page}`).then(function(res) {
        if(!res.ok) {
            res.text().then(text => alert("Failed Downloading songs: " + text))
            return
        }

        res.json().then(function(data) {
            console.log(data)

            for(song of data["Songs"]) {
                songDiv = document.createElement("div")
                songDiv.className = "playlist"

                thumbnail = document.createElement("img")
                thumbnail.className = "thumbnail"
                thumbnail.setAttribute("src", song["ThumbnailURL"])
                songDiv.appendChild(thumbnail)

                detailsDiv = document.createElement("div")
                detailsDiv.className = "details"

                title = document.createElement("div")
                title.className = "title"
                title.innerHTML = song["Title"]
                detailsDiv.appendChild(title)
                songDiv.appendChild(detailsDiv)

                downloadIframe = document.createElement("iframe")
                downloadIframe.setAttribute("style", "display: none;")
                downloadIframe.setAttribute("src", song["DownloadUrl"])
                songDiv.appendChild(downloadIframe)

                document.body.appendChild(songDiv)
            }

            if(data["Next"] != "") {
                download_page(playlist_id, data["Next"])
            }
        })
    })
}