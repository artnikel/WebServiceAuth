$(document).ready(function() {
    $("#updateBalance").click(function() {
      $("#balance").text(balance);
    });

    $("#deposit").click(function() {
      $("#depositModal").modal("show");
    });
  });