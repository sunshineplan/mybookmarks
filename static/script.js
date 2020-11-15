post = (url, obj) => {
  return fetch(url, {
    method: 'post',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: new URLSearchParams(obj)
  }).then(resp => {
    if (resp.status == 401)
      return BootstrapButtons.fire('Error', 'Login status has changed. Please Re-login!', 'error')
        .then(() => window.location = '/')
    else return resp
  })
}

BootstrapButtons = Swal.mixin({
  customClass: {
    confirmButton: 'swal btn btn-primary'
  },
  buttonsStyling: false
});

valid = () => {
  var result = true
  Array.from(document.getElementsByTagName('input'))
    .forEach(i => {
      if (!i.checkValidity())
        result = false
    })
  return result
}
