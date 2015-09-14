function loadRSS() {
	$.get("/rss", function(data) {
		console.log(data);
	});
}
