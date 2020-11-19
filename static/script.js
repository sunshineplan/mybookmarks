BootstrapButtons = Swal.mixin({
  customClass: { confirmButton: 'swal btn btn-primary' },
  buttonsStyling: false
});

valid = () => {
  var result = true
  Array.from(document.getElementsByTagName('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

post = (url, data) => {
  return fetch(url, {
    method: 'post',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .catch(e => {
      return Promise.reject(BootstrapButtons.fire('Error', e, 'error'))
    })
    .then(resp => {
      if (resp.status == 401)
        return BootstrapButtons.fire('Error', 'Login status has changed. Please Re-login!', 'error')
          .then(() => window.location = '/')
      return resp
    })
}

confirm = type => {
  return Swal.fire({
    title: 'Are you sure?',
    text: 'This ' + type + ' will be deleted permanently.',
    icon: 'warning',
    confirmButtonText: 'Delete',
    showCancelButton: true,
    focusCancel: true,
    customClass: {
      confirmButton: 'swal btn btn-danger',
      cancelButton: 'swal btn btn-primary'
    },
    buttonsStyling: false
  }).then(confirm => { return confirm.isConfirmed })
}
