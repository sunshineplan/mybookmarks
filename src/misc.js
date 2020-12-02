import Swal from 'sweetalert2'

export const BootstrapButtons = Swal.mixin({
  customClass: { confirmButton: 'swal btn btn-primary' },
  buttonsStyling: false
});

export const valid = () => {
  var result = true
  Array.from(document.querySelectorAll('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

export const post = async (url, data) => {
  let resp
  try {
    resp = await fetch(url, {
      method: 'post',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
  } catch (e) {
    return Promise.reject(await BootstrapButtons.fire('Error', e, 'error'))
  }
  if (resp.status != 401) return resp
  await BootstrapButtons.fire('Error', 'Login status has changed. Please Re-login!', 'error')
  window.location = '/'
}

export const confirm = async type => {
  const confirm = await Swal.fire({
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
  })
  return confirm.isConfirmed
}
