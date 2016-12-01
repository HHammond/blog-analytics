(function() {
    var img = new Image,
        url = encodeURIComponent(document.location.href),
        title = encodeURIComponent(document.title),
        ref = encodeURIComponent(document.referrer),
        ver = '0.0.1',
        evt = 'sl';
    img.src = '127.0.0.1:8080/a.gif?url=' + url + '&t=' + title + '&ref=' + ref + '&ver=' + ver + '%&evt=' + evt;
})();
