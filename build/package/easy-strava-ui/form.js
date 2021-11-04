$(document).ready(function () {
    $("form").submit(function (event) {
      var formData = {
        name: $("#name").val(),
        type: $("#type").val(),
        start_date_local: $("#start_date_local").val(),
        elapsed_time: $("#elapsed_time").val(),
      };
  
      $.ajax({
        type: "POST",
        url: "http://03d16d1a531c.mylabserver.com:30080/activities",
        data: formData,
        dataType: "json",
        encode: true,
      }).done(function (data) {
        console.log(data);
      });
  
      event.preventDefault();
    });
  });