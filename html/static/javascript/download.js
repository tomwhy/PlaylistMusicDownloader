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

  document.body.appendChild(songDiv);
  
  var downloadLink = document.createElement("a")
  downloadLink.href = song.download_url
  downloadLink.target = "_blank"
  downloadLink.download = `${song.title}.mp3`
  downloadLink.click();
}
