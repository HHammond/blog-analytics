var SimpleAnalytics = (function() {

    function generateGuid() {
        var result, i, j;
        result = '';
        for (j = 0; j < 32; j++) {
            if (j == 8 || j == 12 || j == 16 || j == 20) {
                result = result + '-';
            }
            i = Math.floor(Math.random() * 16).toString(16).toUpperCase();
            result = result + i;
        }
        return result
    }

    const version = '0.0.1';
    const endpoint = 'http://127.0.0.1:8080/a.gif';
    const session = {
        ut: localStorage.getItem('ut') || generateGuid(),
        st: sessionStorage.getItem('st') || generateGuid(),
    };

    sessionStorage.setItem('st', session.st);
    localStorage.setItem('ut', session.ut);

    function buildURI(base, params) {
        var args = [];
        for (key in params) {
            args.push(key + '=' + encodeURIComponent(params[key]));
        }
        return args.length > 0 ? base + '?' + args.join('&') : base;
    }

    var idx = 0;

    function sendPayload(params) {
        params.idx = idx++;

        var img = new Image;
        img.src = buildURI(endpoint, params);
        img.onload = function() {
            img.src = '';
        };
    }

    function getBasePayload() {
        var params = {
            url: encodeURIComponent(document.location.href),
            title: encodeURIComponent(document.title),
            ref: encodeURIComponent(document.referrer),
            ver: version,
        };

        var ut = localStorage.getItem('ut');
        var st = sessionStorage.getItem('st');
        if (ut !== null) {
            params.ut = ut;
        }
        if (st !== null) {
            params.st = st;
        }
        return params
    }

    function sendEvent(type) {
        var params = getBasePayload();
        params.evt = type;
        sendPayload(params);
    }

    function sendPageLoadEvent() {
        sendEvent('sl');
    }

    function sendPageUpdateEvent() {
        sendEvent('up');
    }

    sendPageLoadEvent();
    var sendUpdate = setInterval(sendPageUpdateEvent, 5000);

    return {
        stopUpdates: function() {
            return clearInterval(sendUpdate);
        },
    };
})();
