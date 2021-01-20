<script lang="ts">
  import { fire, post, valid, confirm } from "../misc";
  import { component, bookmark, categories } from "../stores";

  let name = $bookmark ? $bookmark.bookmark : "";
  let url = $bookmark ? $bookmark.url : "";
  let category = $bookmark ? $bookmark.category : "";
  let validated = false;

  $: mode = window.location.pathname == "/" ? "Add" : "Edit";

  const chkURL = () => {
    if (url && !url.match(/^https?:/) && url.length) url = "http://" + url;
  };

  const save = async () => {
    if (valid()) {
      validated = false;
      let resp;
      if (mode == "Add")
        resp = await post("/bookmark/add", {
          name,
          url,
          category,
        });
      else
        resp = await post("/bookmark/edit/" + $bookmark.id, {
          name,
          url,
          category,
        });
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        const json = await resp.json();
        if (json.status == 1) {
          //store
          //back
        } else {
          await fire("Error", json.message, "error");
          if (json.error == 1) name = "";
          else if (json.error == 2) url = "";
        }
      }
    } else validated = true;
  };

  const del = async () => {
    if (await confirm("bookmark")) {
      const resp = await post("/bookmark/delete/" + $bookmark.id);
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        //delete
        //back
      }
    }
  };

  const cancel = async () => {
    window.history.pushState({}, "", "/");
    $component = "show";
  };
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") cancel();
  }}
/>

<svelte:head>
  <title>{mode} Bookmark - My Bookmarks</title>
</svelte:head>

<div on:keydown={save}>
  <header style="padding-left: 20px">
    <h3>{mode} Bookmark</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="form-group">
      <label for="bookmark">Bookmark</label>
      <input
        class="form-control"
        id="bookmark"
        bind:value={name}
        maxlength="40"
        required
      />
      <div class="invalid-feedback">This field is required.</div>
      <small class="form-text text-muted">Max length: 40 characters.</small>
    </div>
    <div class="form-group">
      <label for="url">URL</label>
      <input
        class="form-control"
        id="url"
        type="url"
        bind:value={url}
        on:blur={chkURL}
        required
      />
      <div class="invalid-feedback">Please enter a valid URL.</div>
    </div>
    <div class="form-group">
      <label for="category">Category</label>
      <input
        class="form-control"
        id="category"
        list="category-list"
        bind:value={category}
        maxlength="15"
      />
      <datalist id="category-list">
        {#each $categories as category (category.id)}
          <option>{category.category}</option>
        {/each}
      </datalist>
      <small class="form-text text-muted">
        Max length: 15 characters. One chinese character equal three characters.
      </small>
    </div>
    <button class="btn btn-primary" on:click={save}>{mode}</button>
    <button class="btn btn-primary" on:click={cancel}>Cancel</button>
  </div>
  {#if mode == "Edit"}
    <div class="form">
      <button class="btn btn-danger delete" on:click={del}>Delete</button>
    </div>
  {/if}
</div>
