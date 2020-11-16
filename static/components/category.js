const category = {
  props: { category: Object },
  data() {
    return {
      name: this.category.Name,
      validated: false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px'>
    <a class='h3 title'>{{ mode }} Category</a>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' v-model='name' id='category' maxlength=15 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='save'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='category.ID != undefined'>
    <button class='btn btn-danger delete' @click='del'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Category - My Bookmarks' },
  methods: {
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.category.ID == undefined)
          r = post('/category/add', { category: this.name })
        else
          r = post('/category/edit/' + this.category.ID, { category: this.name })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) this.$parent.content = 'showBookmark'
            else BootstrapButtons.fire('Error', json.message, 'error')
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    del: function () {
      confirm('category').then(confirm => {
        if (confirm) post('/category/delete/' + this.category.ID)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else this.$parent.content = 'showBookmark'
          })
          .catch(e => BootstrapButtons.fire('Error', e, 'error'))
      })
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.category.ID == undefined) return 'Add'
      return 'Edit'
    }
  }
}
