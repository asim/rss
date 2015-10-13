var rssUrl = "/rss";
var chatUrl = "http://malten.me";
var chatPre = "#asl.am.chat.";
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
		var dTime = document.createElement('div');
		d1.className = 'item';
		dTime.className = 'time';
		dTime.innerHTML = parseDate(array[i].Created);
		dTime.setAttribute('data-time', array[i].Created);

		if (array[i].Metadata != null) {
			var a1 = document.createElement('a');
			var a2 = document.createElement('a');
			var a3 = document.createElement('a');
			var a4 = document.createElement('a');
			var dImage = document.createElement('div');
			var dTitle = document.createElement('div');
			var dDesc = document.createElement('div');
			var dBimage = document.createElement('div');
			var dNav = document.createElement('div');	
			var dChat = document.createElement('div');	
			var dTweet = document.createElement('div');	
			
			a3.href = chatUrl + "/" + chatPre + btoa(unescape(encodeURIComponent(array[i].Metadata.Title + array[i].Metadata.Url))).slice(1, 11);
			a3.innerHTML = '<img src="/m.jpg" style="width:24px;height:24px;vertical-align:middle;">';

			a4.href = "http://twitter.com/share?url=/&text=" + encodeURIComponent(array[i].Metadata.Title + " " + array[i].Metadata.Url + " via @_asl_am");
			a4.innerHTML = "<img src=/t.png />";

			dImage.className = 'image';
			dTitle.className = 'title';
			dDesc.className = 'desc';
			dChat.className = "chat"
			dNav.className = "inav"
			dTweet.className = "tweet"

			a1.innerHTML = array[i].Metadata.Title;
			a1.href = array[i].Metadata.Url;
			a2.href = array[i].Metadata.Url;
			dDesc.innerHTML = array[i].Metadata.Description;
			dBimage.style.backgroundImage = "url('" + array[i].Metadata.Image + "')";
			a2.appendChild(dBimage);
			dChat.appendChild(a3);
			dTweet.appendChild(a4);
			dNav.appendChild(dTime);
			dNav.appendChild(dChat);
			dNav.appendChild(dTweet);
			dImage.appendChild(a2);
			dTitle.appendChild(a1);
			d1.appendChild(dTitle);
			d1.appendChild(dDesc);
			d1.appendChild(dNav);
			item.appendChild(dImage);
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

	$('.tweet > a').click(function(event) {
	    var width  = 575,
		height = 400,
		left   = ($(window).width()  - width)  / 2,
		top    = ($(window).height() - height) / 2,
		url    = this.href,
		opts   = 'status=1' +
			 ',width='  + width  +
			 ',height=' + height +
			 ',top='    + top    +
			 ',left='   + left;
	    
	    window.open(url, 'twitter', opts);
	 
	    return false;
	});
};

function loadListeners() {
	if (window.navigator.standalone) {
		$.ajaxSetup({isLocal:true});
	};

        $(window).scroll(function() {
                if($(window).scrollTop() == $(document).height() - $(window).height()) {
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

