function download_songs(playlist_id) {
  ws = new WebSocket(
    `wss://${window.location.host}/api/download/${playlist_id}`
  );

  ws.onmessage = function (msg) {
    console.log(msg.data)
    data = JSON.parse(msg.data);
    console.log(data);

    if (data.type == "error") {
      alert(data.data);
    } else {
      AppendSongToPage(data.data)
    }
  };
}

function AppendSongToPage(song) {
  songDiv = document.createElement("div");
  songDiv.className = "playlist";
  songDiv.setAttribute("name", "song")

  thumbnail = document.createElement("img");
  thumbnail.className = "thumbnail";
  thumbnail.setAttribute("src", song.thumbnail_url);
  songDiv.appendChild(thumbnail);

  detailsDiv = document.createElement("div");
  detailsDiv.className = "details";

  title = document.createElement("div");
  title.className = "title";
  title.innerHTML = song.title;

  detailsDiv.appendChild(title);
  songDiv.appendChild(detailsDiv);
  songDiv.onclick = function(event) {
    browser.downloads.download({
      url: song.download_url,
      filename: `${song.title}.mp3`,
      conflictAction : 'uniquify'
    })


    // var download = document.createElement("iframe")
    // download.setAttribute("style", "display: none;") 
    // download.setAttribute("src", song.download_url)
    // document.body.appendChild(download)
  }

  document.body.appendChild(songDiv);
}

function DownloadAll() {
  for(let song of document.getElementsByName("song")) {
    song.click();
  }
}