const SimpleAnaltyics_version = '0.0.1';
const SimpleAnaltyics_endpoint = '127.0.0.1:8080/a.gif';

var SimpleAnalytics = SimpleAnalytics || {

    generateGUID: function() {
        // From http://stackoverflow.com/a/105074
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
    },

    generateViewKey: function() {
        var result = '';
        for (var j = 0; j < 8; j++) {
            result += Math.floor(Math.random() * 36).toString(36).toUpperCase();
        }
        return result;
    },

    buildURI: function(base, params, protocol) {
        protocol = protocol || document.location.protocol;
        base = protocol + '//' + base;

        var args = [];
        for (key in params) {
            if (params[key] !== null && params[key] != '') {
                args.push(key + '=' + encodeURIComponent(params[key]));
            }
        }
        return args.length > 0 ? base + '?' + args.join('&') : base;
    },

    getVersion: function() {
        return SimpleAnaltyics_version;
    },

    getEndpoint: function() {
        return SimpleAnaltyics_endpoint;
    },

    getParams: function() {
        var params = {
            url: encodeURIComponent(document.location.href),
            title: encodeURIComponent(document.title),
            ref: encodeURIComponent(document.referrer),
            ver: this.getVersion(),
            ut: localStorage.getItem('ut'),
            st: sessionStorage.getItem('st'),
            vwix: this.vwix || '',
        };
        return params
    },

    requestCount: 0,

    sendEvent: function(evt) {
        var params = this.getParams();
        params.idx = this.requestCount++;
        params.evt = evt || '';

        var img = new Image;
        img.src = this.buildURI(this.getEndpoint(), params);
        img.onload = function() {
            this.src = '';
        };
    },

    sendPageLoadEvent: function() {
        return this.sendEvent('in');
    },

    sendPageUpdateEvent: function() {
        return this.sendEvent('up');
    },

    sendPageExitEvent: function() {
        return this.sendEvent('ex');
    },

    startUpdates: function() {
        const updateSeconds = 10
        var sendUpdate = this.sendPageUpdateEvent.bind(this);

        if (this.updates) {
            this.stopUpdates();
        }

        this.updates = setInterval(sendUpdate, updateSeconds * 1000);
        return this.updates;
    },

    stopUpdates: function() {
        clearInterval(this.updates);
        this.updates = null;
    },
};

(function() {
    localStorage.setItem('ut', localStorage.getItem('ut') || SimpleAnalytics.generateGUID());
    sessionStorage.setItem('st', sessionStorage.getItem('st') || SimpleAnalytics.generateGUID());

    SimpleAnalytics.vwix = SimpleAnalytics.generateViewKey();
    SimpleAnalytics.sendPageLoadEvent();
    SimpleAnalytics.startUpdates();

    window.onunload = function() {
        SimpleAnalytics.sendPageExitEvent();
    };
})();
