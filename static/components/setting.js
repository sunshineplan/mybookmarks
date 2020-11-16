const setting = {
  data() {
    return {
      password: '',
      password1: '',
      password2: '',
      validated: false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px'>
    <h3>Setting</h3>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='password'>Current Password</label>
      <input class='form-control' type='password' v-model='password' id='password' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password1'>New Password</label>
      <input class='form-control' type='password' v-model='password1' id='password1' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password2'>Confirm Password</label>
      <input class='form-control' type='password' v-model='password2' id='password2' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max password length: 20 characters.</small>
    </div>
    <button class='btn btn-primary' @click='setting'>Change</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
</div>`,
  mounted() { document.title = 'Setting - My Bookmarks' },
  methods: {
    setting: function () {
      if (valid()) {
        this.validated = false
        post('/setting', {
          password: this.password,
          password1: this.password1,
          password2: this.password2
        }).then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1)
              BootstrapButtons.fire('Success', 'Your password has changed. Please Re-login!', 'success')
                .then(() => window.location = '/')
            else
              BootstrapButtons.fire('Error', json.message, 'error')
                .then(() => {
                  if (json.error == 1) this.password = ''
                  else {
                    this.password1 = ''
                    this.password2 = ''
                  }
                })
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  }
}
