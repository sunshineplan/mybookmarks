$(document).on('click', '.toggle', () => $('.sidebar').toggle('slide'))

$(document).on('click', '.content', () => {
  if ($('.sidebar').is(':visible') && $(window).width() <= 900)
    $('.sidebar').toggle('slide')
});
