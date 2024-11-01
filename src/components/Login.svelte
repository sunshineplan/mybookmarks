<script lang="ts">
  import { encrypt, fire, post } from "../misc.svelte";

  let { info }: { info: () => Promise<void> } = $props();

  let username = $state(localStorage.getItem("username"));
  let password = $state("");
  let rememberme = $state(
    localStorage.getItem("rememberme") == "true" ? true : false,
  );
  let usernameInput: HTMLInputElement;
  let passwordInput: HTMLInputElement;

  const login = async () => {
    if (!usernameInput.checkValidity())
      await fire("Error", "Username cannot be empty.", "error");
    else if (!passwordInput.checkValidity())
      await fire("Error", "Password cannot be empty.", "error");
    else {
      var pwd: string;
      if (window.pubkey && window.pubkey.length)
        pwd = encrypt(window.pubkey, password);
      else pwd = password;
      const resp = await post(
        window.universal + "/login",
        { username, password: pwd, rememberme },
        true,
      );
      if (resp.ok) {
        const json = await resp.json();
        if (json.status == 1) {
          if (rememberme) localStorage.setItem("rememberme", "true");
          else localStorage.removeItem("rememberme");
          await info();
        } else await fire("Error", json.message, "error");
      } else await fire("Error", await resp.text(), "error");
    }
  };
</script>

<svelte:head>
  <title>Log In - My Bookmarks</title>
</svelte:head>

<div class="content">
  <header>
    <h3
      class="d-flex justify-content-center align-items-center"
      style="height: 100%"
    >
      Log In
    </h3>
  </header>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="login"
    onkeydown={async (e) => {
      if (e.key == "Enter") await login();
    }}
  >
    <div class="mb-3">
      <label class="form-label" for="username">Username</label>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="form-control"
        bind:this={usernameInput}
        bind:value={username}
        id="username"
        maxlength="20"
        placeholder="Username"
        autofocus
        required
      />
    </div>
    <div class="mb-3">
      <label class="form-label" for="password">Password</label>
      <input
        class="form-control"
        type="password"
        bind:this={passwordInput}
        bind:value={password}
        id="password"
        maxlength="20"
        placeholder="Password"
        required
      />
    </div>
    <div class="form-group form-check">
      <input
        type="checkbox"
        class="form-check-input"
        bind:checked={rememberme}
        id="rememberme"
      />
      <label class="form-check-label" for="rememberme">Remember Me</label>
    </div>
    <hr />
    <button class="btn btn-primary login" onclick={login}>Log In</button>
  </div>
</div>

<style>
  .login {
    width: 250px;
    margin: 0 auto;
  }

  .form-control {
    width: 250px;
  }
</style>
