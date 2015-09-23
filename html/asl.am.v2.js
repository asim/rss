var rssUrl = "/rss";
var limit = 25;
var last = timeAgo();
var typing = false;
var maxChars = 500;
var maxThoughts = 1000;
var seen = {};
var streams = {};

String.prototype.parseURL = function() {
	return this.replace(/[A-Za-z]+:\/\/[A-Za-z0-9-_]+\.[A-Za-z0-9-_:%&~\?\/.=]+/g, function(url) {
		var pretty = url.replace(/^http(s)?:\/\/(www\.)?/, '');
		return pretty.link(url);
	});
};

function timeAgo() {
	var ts = new Date().getTime() / 1000;
	return (ts - 86400) * 1e9;
};

function parseDate(tdate) {
    var system_date = new Date(tdate/1e6);
    var user_date = new Date();
    if (K.ie) {
        system_date = Date.parse(tdate.replace(/( \+)/, ' UTC$1'))
    }
    var diff = Math.floor((user_date - system_date) / 1000);
    if (diff < 0) {return "0s";}
    if (diff < 60) {return diff + "s";}
    if (diff <= 90) {return "1m";}
    if (diff <= 3540) {return Math.round(diff / 60) + "m";}
    if (diff <= 5400) {return "1h";}
    if (diff <= 86400) {return Math.round(diff / 3600) + "h";}
    if (diff <= 129600) {return "1d";}
    if (diff < 604800) {return Math.round(diff / 86400) + "d";}
    if (diff <= 777600) {return "1w";}
    return "on " + system_date;
};

// from http://widgets.twimg.com/j/1/widget.js
var K = function () {
    var a = navigator.userAgent;
    return {
        ie: a.match(/MSIE\s([^;]*)/)
    }
}();

function escapeHTML(str) {
	var div = document.createElement('div');
	div.style.display = 'none';
	div.appendChild(document.createTextNode(str));
	return div.innerHTML;
};

function displayItems(array, direction) {
	var list = document.getElementById('items');

        for(i = 0; i < array.length; i++) {
		if (array[i].Id in seen) {
			continue;
		};

                var item = document.createElement('li');
		var html = escapeHTML(array[i].Text);
		var d1 = document.createElement('div');
		var d2 = document.createElement('div');
		d1.className = 'item';
		d2.className = 'time';
		d2.innerHTML = parseDate(array[i].Created);
		d2.setAttribute('data-time', array[i].Created);

		if (array[i].Metadata != null) {
			var a1 = document.createElement('a');
			var a2 = document.createElement('a');
			var d3 = document.createElement('div');
			var d4 = document.createElement('div');
			var d5 = document.createElement('div');
			var d6 = document.createElement('div');

			a1.innerHTML = array[i].Metadata.Title;
			a1.href = array[i].Metadata.Url;
			a2.href = array[i].Metadata.Url;
			d3.className = 'image';
			d4.className = 'title';
			d5.className = 'desc';
			d6.style.backgroundImage = "url('" + array[i].Metadata.Image + "')";
			a2.appendChild(d6);
			d3.appendChild(a2);
			d4.appendChild(a1);
			d5.innerHTML = array[i].Metadata.Description;
			d1.appendChild(d4);
			d1.appendChild(d5);
			d1.appendChild(d2);
			item.appendChild(d3);
			item.appendChild(d1);
			if (direction >= 0) {
				list.insertBefore(item, list.firstChild);
			} else {
				list.appendChild(item);
			}
		};

		seen[array[i].Id] = array[i];
        }

	if (direction >= 0) {
		last = array[array.length -1].Created;
	}
};

function loadListeners() {
	if (window.navigator.standalone) {
		$.ajaxSetup({isLocal:true});
	};

	$(".rss").scroll(function() {
		if($(this).scrollTop() + $(this).innerHeight() >= this.scrollHeight) {
			console.log("hit bottom");
			loadMore();
		}
	});
};

function loadMore() {
	var divs = document.getElementsByClassName('time');
	var oldest = new Date().getTime() * 1e6;

	if (divs.length > 0) {
		oldest = divs[divs.length-1].getAttribute('data-time');
	}

	var params = "?stream=tech&direction=-1&limit=" + limit + "&last=" + oldest;

	$.get(rssUrl + params, function(data) {
		if (data != undefined && data.length > 0) {
			displayItems(data, -1);
		}
	})
	.fail(function(err) {
		console.log(err);
	})
	.done();

        return false;
};

function loadRSS() {
	var params = "?stream=tech&direction=1&limit=" + limit + "&last=" + last;

	$.get(rssUrl + params, function(data) {
		if (data != undefined && data.length > 0) {
			displayItems(data, 1);
	    		updateTimestamps();
		}
	})
	.fail(function(err) {
		console.log(err);
	})
	.done();

        return false;
};

function pollRSS() {
        loadRSS();

        setTimeout(function() {
            pollRSS();
        }, 60000);
};

function pollTimestamps() {
        updateTimestamps();

        setTimeout(function() {
            pollTimestamps();
        }, 60000);
};

function updateTimestamps() {
	var divs = document.getElementsByClassName('time');
	for (i = 0; i < divs.length; i++) {
		var time = divs[i].getAttribute('data-time');
		divs[i].innerHTML = parseDate(time);
	};
};
