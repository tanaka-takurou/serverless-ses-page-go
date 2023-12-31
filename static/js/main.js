$(document).ready(function() {
  GetMails();
});

var GetMails = function() {
  const data = {action: "getlist"};
  request(data, (res)=>{
    $("#mails div").remove()
    res.data.forEach(v => {
      var nameTag = $("<div></div>", {
        "class": "header"
      }).text(v.subject);
      var dateTag = $("<span></span>", {
        "class": "date"
      }).text(v.date);
      var descriptionTag = $("<div></div>", {
        "class": "description"
      }).text(v.from).append(dateTag);
      var contentTag = $("<div></div>", {
        "class": "content"
      }).append(nameTag).append(descriptionTag);
      var buttonTag = $("<div></div>", {
        "class": "ui primary button",
        "onclick": "GetMail(\"" + v.file + "\");"
      }).text("Show");
      var floatedTag = $("<div></div>", {
        "class": "right floated content"
      }).append(buttonTag);
      var iconTag = $("<i></i>", {
        "class": "large envelope outline middle aligned icon"
      });
      var itemTag = $("<div></div>", {
        "class": "item"
      }).append(floatedTag).append(iconTag).append(contentTag);
      $("#mails").append(itemTag);
    });
  }, (e)=>{
    console.log(e.responseJSON.message);
    $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
  });
};

var GetMail = function(fileName) {
  const data = {action: "getbody", name: fileName};
  request(data, (res)=>{
    if (!!res && !!res.message && res.message.length > 0) {
      $("#result").text(res.message);
      $("#info").removeClass("hidden").addClass("visible");
    }
  }, (e)=>{
    console.log(e.responseJSON.message);
    $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
  });
};

var request = function(data, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};
var App = { url: location.origin + {{ .ApiPath }} };
