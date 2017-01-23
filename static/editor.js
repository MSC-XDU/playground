/**
 * Created by tuzi on 2017/1/22.
 */

var Editors = function (options) {
    // FIXME: 为了偷懒，css 直接放在这里。
    var css = ".playground-editor input[type=submit]{-moz-box-shadow:inset 0 1px 0 0 #7a8eb9;-webkit-box-shadow:inset 0 1px 0 0 #7a8eb9;box-shadow:inset 0 1px 0 0 #7a8eb9;background:-webkit-gradient(linear,left top,left bottom,color-stop(0.05,#637aad),color-stop(1,#5972a7));background:-moz-linear-gradient(top,#637aad 5%,#5972a7 100%);background:-webkit-linear-gradient(top,#637aad 5%,#5972a7 100%);background:-o-linear-gradient(top,#637aad 5%,#5972a7 100%);background:-ms-linear-gradient(top,#637aad 5%,#5972a7 100%);background:linear-gradient(to bottom,#637aad 5%,#5972a7 100%);filter:progid:DXImageTransform.Microsoft.gradient(startColorstr='#637aad',endColorstr='#5972a7',GradientType=0);background-color:#637aad;border:1px solid #314179;display:inline-block;cursor:pointer;color:#fff;font-family:Arial;font-size:13px;font-weight:bold;padding:6px 12px;text-decoration:none;outline:0;margin:0 0 0 auto}.playground-editor input[type=submit]:hover{background:-webkit-gradient(linear,left top,left bottom,color-stop(0.05,#5972a7),color-stop(1,#637aad));background:-moz-linear-gradient(top,#5972a7 5%,#637aad 100%);background:-webkit-linear-gradient(top,#5972a7 5%,#637aad 100%);background:-o-linear-gradient(top,#5972a7 5%,#637aad 100%);background:-ms-linear-gradient(top,#5972a7 5%,#637aad 100%);background:linear-gradient(to bottom,#5972a7 5%,#637aad 100%);filter:progid:DXImageTransform.Microsoft.gradient(startColorstr='#5972a7',endColorstr='#637aad',GradientType=0);background-color:#5972a7}.playground-editor input[type=submit]:active{position:relative;top:1px}.playground-editor .header{display:flex;flex-direction:row;justify-content:flex-start;align-items:center;padding:5px;background-color:#dde8f2}.playground-editor .header .title{display:inline-block;margin:0}.playground-editor .run-result pre{padding: 0 5px 0 5px;white-space:pre-wrap;overflow-y:scroll;overflow-x:hidden}";
    var all = document.querySelectorAll(".playground-editor");
    var editors = {};
    options = options || {};

    function codeUpload(e) {
        e.preventDefault();
        var f = e.target;
        var resultEcho = f.querySelector(".run-result pre");
        resultEcho.innerText = "等待服务器相应...";
        editors[f.dataset.id].save();
        var data = new FormData(f);

        var http = new XMLHttpRequest();
        http.open("POST", f.dataset.url);
        http.onreadystatechange = function () {
            if (http.readyState == 4) {
                resultEcho.innerText = http.responseText;
            }
        };
        http.send(data)
    }

    function buildForm(el) {
        var textarea = document.createElement("textarea");
        textarea.setAttribute("name", "code");
        var code = el.querySelector("code");
        if (code) {
            textarea.value = code.textContent;
            el.removeChild(code);
        }

        var header = document.createElement("div");
        header.className = "header";
        var title = document.createElement("h2");
        title.className = "title";
        title.textContent = "MSC-playground";
        header.appendChild(title);

        var form = document.createElement("form");
        form.dataset.url = el.dataset.url;
        form.dataset.id = el.dataset.id;

        var submit = document.createElement("input");
        submit.setAttribute("type", "submit");
        submit.setAttribute("value", "运行");
        header.appendChild(submit);

        var resultDiv = document.createElement("div");
        resultDiv.className = "run-result";
        var result = document.createElement("pre");
        if (options.resultLines) {
            result.style["max-height"] = options.resultLines + 'em';
            result.style["min-height"] = options.resultLines + 'em';
        }
        resultDiv.appendChild(result);

        form.appendChild(header);
        form.appendChild(textarea);
        form.appendChild(resultDiv);
        el.appendChild(form);

        form.addEventListener("submit", codeUpload);

        return textarea
    }

    // FIXME: 偷懒使用了阻塞调用。
    function syncGet(url, callback) {
        var http = new XMLHttpRequest();
        http.open("GET", url, false);
        http.onreadystatechange = function () {
            if (http.readyState == 4) {
                callback(http.responseText)
            }
        };
        http.send();
    }

    function injectCSS() {
        var appendStyle = function (css) {
            var style = document.createElement("style");
            style.textContent = css;
            document.head.appendChild(style);
        };

        var loadCSS = function (url) {
            syncGet(url, appendStyle)
        };

        appendStyle(css);
        loadCSS("https://cdn.bootcss.com/codemirror/5.23.0/codemirror.min.css");
        loadCSS("https://cdn.bootcss.com/codemirror/5.23.0/theme/solarized.min.css");
    }

    function setup() {
        Array.prototype.forEach.call(all, function (el) {
            editors[el.dataset.id] = CodeMirror.fromTextArea(buildForm(el), {
                mode: el.dataset.lang,
                theme: "solarized light",
                lineNumbers: true
            })
        });
        options.done && options.done();
    }

    if (options.autoInject) {
        injectCSS();
        if (!window.$LAB) {
            syncGet("https://cdn.bootcss.com/labjs/2.0.3/LAB.min.js", function (content) {
                var script = document.createElement("script");
                script.textContent = content;
                script.type = "text/javascript";
                document.head.appendChild(script);
            });
        }

        var lab = $LAB.script("https://cdn.bootcss.com/codemirror/5.23.0/codemirror.min.js").wait();
        Array.prototype.forEach.call(all, function (el) {
            var lang = {};
            var l = el.dataset.lang;
            if (!lang[l]) {
                lab = lab.script("https://cdn.bootcss.com/codemirror/5.23.0/mode/" + l + "/" + l + ".min.js");
                lang[l] = true;
            }
        });
        lab.wait(setup)
    } else {
        setup();
    }

    this.editors = editors;
    this.options = options;
};