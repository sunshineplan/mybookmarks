<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, post } from "../misc";
  import { component } from "../stores";

  const dispatch = createEventDispatcher();

  let username = localStorage.getItem("username") as string;
  let password = "";
  let rememberme = false;

  const login = async () => {
    if (
      !(document.querySelector(
        "#username"
      ) as HTMLSelectElement).checkValidity()
    )
      await fire("Error", "Username cannot be empty.", "error");
    else if (
      !(document.querySelector(
        "#password"
      ) as HTMLSelectElement).checkValidity()
    )
      await fire("Error", "Password cannot be empty.", "error");
    else {
      const resp = await post("/login", {
        username,
        password,
        rememberme,
      });
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        localStorage.setItem("username", username);
        dispatch("info");
        window.history.pushState({}, "", "/");
        $component = "show";
      }
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
    >Log In</h3>
  </header>
  <div
    class="login"
    on:keydown={async (e) => {
      if (e.key == "Enter") await login();
    }}
  >
    <div class="form-group">
      <label for="username">Username</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        class="form-control"
        bind:value={username}
        id="username"
        maxlength="20"
        placeholder="Username"
        autofocus
        required
      />
    </div>
    <div class="form-group">
      <label for="password">Password</label>
      <input
        class="form-control"
        type="password"
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
    <button class="btn btn-primary login" on:click={login}>Log In</button>
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
