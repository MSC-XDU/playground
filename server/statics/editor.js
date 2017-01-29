var baseURL = location.protocol + "//" + location.hostname +
    (location.port && ":" + location.port);

var urlMaps = {
    "Go": "go",
    "Python": "py",
    "C": "c"
};

var modeMaps = {
    "Go": "text/x-go",
    "Python": "python",
    "C": "text/x-csrc"
};

var defaultMode = document.getElementById("default-mode").textContent;

var editor = CodeMirror.fromTextArea(document.getElementById("code"), {
    theme: "monokai",
    mode: modeMaps[defaultMode],
    lineNumbers: true
});

document.getElementById("run-code").addEventListener("click", function (e) {
    editor.save();
    var code = editor.getTextArea().value;
    runCode(code, currentURL())
});

document.getElementById("share-code").addEventListener("click", function (e) {
    editor.save();
    var code = editor.getTextArea().value;
    shareCode(code, currentURL())
});

var modeSelect = document.getElementById("mode-select");
modeSelect != null && modeSelect.addEventListener("change", function (e) {
    document.getElementById("result").textContent = "";
    editor.setValue("");
    editor.setOption("mode", currentModeName())
});

document.getElementById("menu-bar").addEventListener("click", function (e) {
    var btns = document.getElementById("btns");
    if (btns.style.visibility == "hidden") {
        showMenu();
    } else {
        hideMenu();
    }
});

function runCode(code, type) {
    var data = new FormData();
    data.append("code", code);
    var result = document.getElementById("result");
    result.textContent = "正在连接服务器\n因为服务器性能不足，一次运行约需等待 1 - 3 秒，请耐心等待";

    var http = new XMLHttpRequest();
    http.open("POST", baseURL + "/run/" + type);
    http.onreadystatechange = function () {
        if (http.readyState == 4 && http.status >= 200 && http.status < 300) {
            hideMenu();
            result.textContent = http.responseText
        } else {
            result.textContent = "服务器错误"
        }
    };
    http.send(data)
}

function shareCode(code, type) {
    var data = new FormData();
    data.append("code", code);
    data.append("type", type);

    var result = document.getElementById("result");
    result.textContent = "正在连接服务器";

    var http = new XMLHttpRequest();
    http.open("POST", baseURL + "/share");
    http.responseType = "json";
    http.onreadystatechange = function () {
        if (http.readyState == 4 && http.status >= 200 && http.status < 300) {
            result.textContent = "该代码的分享连接是   " + baseURL + "/share/" + http.response["url"] + "\n快去给别人展示一下";
        } else {
            result.textContent = http.statusText;
        }
    };
    http.send(data)
}

function currentModeName() {
    var select = document.getElementById("mode-select");
    var mode = select ? select.value : defaultMode;
    return modeMaps[mode]
}

function currentURL() {
    var select = document.getElementById("mode-select");
    var mode = select ? select.value : defaultMode;
    return urlMaps[mode]
}

function hideMenu() {
    var btns = document.getElementById("btns");
    btns.style.visibility = "hidden"
}

function showMenu() {
    var btns = document.getElementById("btns");
    btns.style.visibility = "visible"
}
