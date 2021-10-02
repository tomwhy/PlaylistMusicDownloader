function download_songs(playlist_id) {
  ws = new WebSocket(
    `wss://${window.location.host}/api/download/${playlist_id}`
  );

  ws.onmessage = function (msg) {
    console.log(msg.data);
    data = JSON.parse(msg.data);
    console.log(data);

    if (data.type == "error") {
      alert(data.data);
    } else {
      AppendSongToPage(data.data);
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

  title = document.createElement("a");
  title.className = "title";
  title.href = song.download_url;
  title.download = `${song.title}.mp3`;
  title.target = "_blank";
  title.innerHTML = song.title;
  title.setAttribute("name", "song");

  detailsDiv.appendChild(title);
  songDiv.appendChild(detailsDiv);

  document.body.appendChild(songDiv);
}

function DownloadAll() {
  for (let song of document.getElementsByName("song")) {
    song.click();
  }
}
