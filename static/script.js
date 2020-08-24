BootstrapButtons = Swal.mixin({
  customClass: {
    confirmButton: 'swal btn btn-primary'
  },
  buttonsStyling: false
});

isMobile = window.matchMedia('(max-width: 700px)');

isMobile.onchange = simplify_url;

$(document).on('click', '.category', function () {
  $('.category').removeClass('active');
  $(this).addClass('active');
  var categoryID = $(this).prop('id');
  document.cookie = 'LastVisit=' + categoryID + '; Path=/';
  loadBookmarks(categoryID);
  if ($(window).width() <= 900) $('.sidebar').toggle('slide');
});

$(document).on('blur', 'input[type="url"]', function () {
  var string = $(this).val();
  if (!string.match(/^https?:/) && string.length) {
    string = 'http://' + string;
    $(this).val(string);
  };
});

$(document).on('keyup', event => {
  if (event.key == 'ArrowUp') arrow('up');
  else if (event.key == 'ArrowDown') arrow('down');
  else if (event.key == 'Enter') $('#submit').click();
});

$(document).on('click', '.toggle', () => $('.sidebar').toggle('slide'));

$(document).on('click', '.content', () => {
  if ($('.sidebar').is(':visible') && $(window).width() <= 900)
    $('.sidebar').toggle('slide');
});

function load(categoryID) {
  if (categoryID === undefined) categoryID = document.cookie.split('LastVisit=')[1];
  else document.cookie = 'LastVisit=' + categoryID + '; Path=/';
  $.getJSON('/category/get', json => {
    $('#categories').empty();
    $('#-1.category').text('All Bookmarks (' + json.total + ')');
    $.each(json.categories, (index, i) =>
      $('#categories').append("<li><a class='nav-link category' id='" + i.ID + "'>" + i.Name + ' (' + i.Count + ')' + '</a></li>'));
    $('#categories').append("<li><a class='nav-link category' id=0>Uncategorized (" + json.uncategorized + ')' + '</a></li>');
    $('#' + categoryID).addClass('active');
  }).done(() => loadBookmarks(categoryID))
    .fail(jqXHR => checkXHR(jqXHR));
};

function loadBookmarks(categoryID = -1) {
  var param, promise;
  if (categoryID == -1) param = '';
  else param = '?category=' + categoryID;
  loading();
  if (!$('#mybookmarks').length) promise = $.get('/bookmark', html => $('.content').html(html));
  else promise = Promise.resolve();
  promise.then(() => {
    $('tbody').empty();
    $.getJSON('/bookmark/get' + param, json => {
      $.each(json.bookmarks, (i, bookmark) => {
        var $tr = $("<tr data-id='" + bookmark.ID + "'></tr>");
        $tr.append('<td>' + bookmark.Name + '</td>');
        $tr.append("<td><a href='" + bookmark.URL + "' target='_blank' class='url'></a></td>");
        $tr.append('<td>' + bookmark.Category + '</td>');
        $tr.append("<td><a class='icon' onclick='bookmark(" + bookmark.ID + ")'><i class='material-icons edit'>edit</i></a></td>");
        $tr.appendTo('tbody');
      });
      simplify_url();
      document.title = json.category.name + ' - My Bookmarks';
      $('.title').text(json.category.name);
      if (json.category.id > 0) $('#editCategory').show().attr('onclick', 'category(' + json.category.id + ')');
      else $('#editCategory').hide().removeAttr('onclick');
      $('#addBookmark').attr('onclick', 'bookmark(0, ' + json.category.id + ')');
    }).done(() => loading(false))
      .fail(jqXHR => checkXHR(jqXHR));
  }).catch(jqXHR => checkXHR(jqXHR));
};

function category(categoryID = 0) {
  var url;
  if (categoryID == 0) {
    url = '/category/add';
    if ($(window).width() <= 900) $('.sidebar').toggle('slide');
  } else url = '/category/edit/' + categoryID;
  $('.category').removeClass('active');
  loading();
  $.get(url, html => {
    loading(false);
    $('.content').html(html);
    document.title = $('.title').text() + ' - My Bookmarks';
    $('#category').focus();
  }).fail(jqXHR => checkXHR(jqXHR));
};

function bookmark(id = 0, categoryID = 0) {
  var url;
  if (id == 0) {
    if (categoryID > 0) url = '/bookmark/add?category=' + categoryID;
    else url = '/bookmark/add';
  } else url = '/bookmark/edit/' + id;
  $('.category').removeClass('active');
  loading();
  $.get(url, html => {
    loading(false);
    $('.content').html(html);
    document.title = $('.title').text() + ' - My Bookmarks';
    $('#bookmark').focus();
  }).fail(jqXHR => checkXHR(jqXHR));
};

function setting() {
  $('.category').removeClass('active');
  loading();
  $.get('/auth/setting', html => {
    loading(false);
    $('.content').html(html);
    document.title = 'Setting - My Bookmarks';
    $('#password').focus();
  }).fail(jqXHR => checkXHR(jqXHR));
};

function doCategory(id) {
  var url;
  if (id == 0) url = '/category/add';
  else url = '/category/edit/' + id;
  if (valid())
    $.post(url, $('input').serialize(), json => {
      $('.form').removeClass('was-validated');
      if (json.status == 0)
        BootstrapButtons.fire('Error', json.message, 'error').then(() => {
          if (json.error == 1) $('#category').val('');
        });
      else load();
    }).fail(jqXHR => checkXHR(jqXHR));
};

function doBookmark(id) {
  var url;
  if (id == 0) url = '/bookmark/add';
  else url = '/bookmark/edit/' + id;
  if (valid())
    $.post(url, $('input').serialize(), json => {
      $('.form').removeClass('was-validated');
      if (json.status == 0)
        BootstrapButtons.fire('Error', json.message, 'error').then(() => {
          if (json.error == 1) {
            $('#bookmark').val('');
          } else if (json.error == 2) {
            $('#url').val('');
          } else if (json.error == 3) {
            $('#category').val('');
          };
        });
      else goback();
    }).fail(jqXHR => checkXHR(jqXHR));
};

function doDelete(mode, id) {
  var url;
  if (mode == 'category') url = '/category/delete/' + id;
  else if (mode == 'bookmark') url = '/bookmark/delete/' + id;
  else return false;
  Swal.fire({
    title: 'Are you sure?',
    text: 'This ' + mode + ' will be deleted permanently.',
    icon: 'warning',
    confirmButtonText: 'Delete',
    showCancelButton: true,
    focusCancel: true,
    customClass: {
      confirmButton: 'swal btn btn-danger',
      cancelButton: 'swal btn btn-primary'
    },
    buttonsStyling: false
  }).then(confirm => {
    if (confirm.isConfirmed)
      $.post(url, json => {
        if (json.status == 1)
          if (mode == 'bookmark') goback(); else load(-1);
      }).fail(jqXHR => checkXHR(jqXHR));
  });
};

function doSetting() {
  if (valid())
    $.post('/auth/setting', $('input').serialize(), json => {
      $('.form').removeClass('was-validated');
      if (json.status == 1)
        BootstrapButtons.fire('Success', 'Your password has changed. Please Re-login!', 'success')
          .then(() => window.location = '/auth/login');
      else BootstrapButtons.fire('Error', json.message, 'error').then(() => {
        if (json.error == 1)
          $('#password').val('');
        else if (json.error == 2) {
          $('#password1').val('');
          $('#password2').val('');
        };
      });
    }).fail(jqXHR => checkXHR(jqXHR));
};

function checkXHR(xhr) {
  if (xhr.status == 401) window.location = '/auth/login';
};

function valid() {
  var result = true
  $('input').each(function () {
    if ($(this)[0].checkValidity() === false) {
      $('.form').addClass('was-validated');
      result = false;
    };
  });
  return result;
};

function simplify_url() {
  if (isMobile.matches)
    $('.url').each(function () { $(this).text($(this).attr('href').replace(/https?:\/\/(www\.)?/i, '')) });
  else $('.url').each(function () { $(this).text($(this).attr('href')) });
};

function goback() {
  var last = document.cookie.split('LastVisit=')[1];
  $('#' + last).addClass('active');
  loadBookmarks(last);
};

function loading(show = true) {
  if (show) {
    $('.loading').css('display', 'flex');
    $('.content').css('opacity', 0.5);
  } else {
    $('.loading').hide();
    $('.content').css('opacity', 1);
  }
};

function arrow(direction) {
  var index, len = $('.category').length;
  $('.category').each((i, e) => {
    if (e.classList.contains('active')) {
      index = i;
      return false;
    };
  });
  if (direction == 'up')
    if (index > 0) {
      $('.category').removeClass('active');
      $('#' + $('.category')[index - 1].id).addClass('active');
      loadBookmarks($('.category')[index - 1].id);
    };
  if (direction == 'down')
    if (index < len - 1) {
      $('.category').removeClass('active');
      $('#' + $('.category')[index + 1].id).addClass('active');
      loadBookmarks($('.category')[index + 1].id);
    }
};
