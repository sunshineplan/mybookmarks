const login = {
  data() {
    return {
      username: '',
      password: '',
      rememberme: false,
      validated: false
    }
  },
  template: `
  <div @keyup.enter='login'>
    <header>
      <h3 class='d-flex justify-content-center align-items-center' style='height: 100%'>Log In</h3>
    </header>
    <div class='login' :class="{ 'was-validated': validated }">
      <div class='form-group'>
        <label for='username'>Username</label>
        <input autofocus class='form-control' v-model='username' id='username' maxlength=20 placeholder='Username' required>
      </div>
      <div class='form-group'>
        <label for='password'>Password</label>
        <input class='form-control' type='password' v-model='password' id='password' maxlength=20 placeholder='Password' required>
      </div>
      <div class='form-group form-check'>
        <input type='checkbox' class='form-check-input' v-model='rememberme' id='rememberme'>
        <label class='form-check-label' for='rememberme'>Remember Me</label>
      </div>
      <hr>
      <button class='btn btn-primary login' @click='login'>Log In</button>
    </div>
  </div>`,
  mounted() { document.title = 'Log In - My Bookmarks' },
  methods: {
    login: function () {
      if (valid()) {
        this.validated = false
        post('/login', {
          username: this.username,
          password: this.password,
          rememberme: this.rememberme
        }).then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else window.location = '/'
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    }
  }
}
