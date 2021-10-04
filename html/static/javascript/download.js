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
  songDiv.setAttribute("name", "song");
  songDiv.setAttribute("id", song.id)

  thumbnail = document.createElement("img");
  thumbnail.className = "thumbnail";
  thumbnail.setAttribute("src", song.thumbnail_url);
  songDiv.appendChild(thumbnail);

  detailsDiv = document.createElement("div");
  detailsDiv.className = "details";

  title = document.createElement("div");
  title.className = "title"
  title.innerHTML = song.title;
  detailsDiv.appendChild(title);

  audio = document.createElement("audio");
  audio.setAttribute("controls", undefined);

  audioSource = document.createElement("source");
  audioSource.src = song.download_url;
  audioSource.type = "audio/mpeg"
  audio.appendChild(audioSource);
  detailsDiv.appendChild(audio);

  songDiv.appendChild(detailsDiv);

  document.body.appendChild(songDiv);
}

function DownloadAll() {
  for (let song of document.getElementsByTagName("source")) {
    iframe = document.createElement("iframe");
    iframe.style = "display:none"
    iframe.src = song.src
    document.body.appendChild(iframe);
  }
}
