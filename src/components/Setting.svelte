<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { encrypt, fire, post, valid } from "../misc";
  import { component } from "../stores";

  const dispatch = createEventDispatcher();

  let password = "";
  let password1 = "";
  let password2 = "";
  let validated = false;

  const setting = async () => {
    if (valid()) {
      validated = false;
      var pwd: string, p1: string, p2: string;
      if (window.pubkey && window.pubkey.length) {
        pwd = <string>encrypt(window.pubkey, password);
        p1 = <string>encrypt(window.pubkey, password1);
        p2 = <string>encrypt(window.pubkey, password2);
      } else {
        pwd = password;
        p1 = password1;
        p2 = password2;
      }
      const resp = await post(
        window.universal + "/chgpwd",
        {
          password: pwd,
          password1: p1,
          password2: p2,
        },
        true
      );
      if (resp.ok) {
        const json = await resp.json();
        if (json.status == 1) {
          await fire(
            "Success",
            "Your password has changed. Please Re-login!",
            "success"
          );
          dispatch("reload");
          window.history.pushState({}, "", "/");
          $component = "show";
        } else {
          await fire("Error", json.message, "error");
          if (json.error == 1) password = "";
          else {
            password1 = "";
            password2 = "";
          }
        }
      } else await fire("Error", await resp.text(), "error");
    } else validated = true;
  };

  const cancel = () => {
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
  <title>Setting - My Bookmarks</title>
</svelte:head>

<header style="padding-left: 20px">
  <h3>Setting</h3>
  <hr />
</header>
<div
  style="padding-left: 20px"
  class="was-validated: {validated}"
  on:keyup={async (e) => {
    if (e.key == "Enter") await setting();
  }}
>
  <div class="mb-3">
    <label class="form-label" for="password">Current Password</label>
    <input
      class="form-control"
      type="password"
      bind:value={password}
      id="password"
      maxlength="20"
      required
    />
    <div class="invalid-feedback">This field is required.</div>
  </div>
  <div class="mb-3">
    <label class="form-label" for="password1">New Password</label>
    <input
      class="form-control"
      type="password"
      bind:value={password1}
      id="password1"
      maxlength="20"
      required
    />
    <div class="invalid-feedback">This field is required.</div>
  </div>
  <div class="mb-3">
    <label class="form-label" for="password2">Confirm Password</label>
    <input
      class="form-control"
      type="password"
      bind:value={password2}
      id="password2"
      maxlength="20"
      required
    />
    <div class="invalid-feedback">This field is required.</div>
    <small class="form-text text-muted">
      Max password length: 20 characters.
    </small>
  </div>
  <button class="btn btn-primary" on:click={setting}>Change</button>
  <button class="btn btn-primary" on:click={cancel}>Cancel</button>
</div>

<style>
  .form-control {
    width: 250px;
  }
</style>
