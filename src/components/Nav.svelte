<script lang="ts">
  import { component, showSidebar } from "../stores";

  export let username: string;
</script>

<nav class="navbar navbar-light topbar">
  <div class="d-flex" style="height:100%">
    <span
      class="brand"
      on:click={() => {
        window.history.pushState({}, "", "/");
        $component = "show";
      }}> My Bookmarks </span>
  </div>
  <div class="navbar-nav flex-row">
    {#if username}
      <span class="nav-link">{username}</span>
      <span
        class="nav-link link"
        on:click={() => {
          window.history.pushState({}, "", "/setting");
          if (window.innerWidth <= 900) showSidebar.close();
          $component = "setting";
        }}> Setting </span>
      <a class="nav-link link" href="/logout">Log out</a>
    {:else}
      <span class="nav-link">Log in</span>
    {/if}
  </div>
</nav>

<style>
  .topbar {
    position: fixed;
    top: 0px;
    z-index: 2;
    width: 100%;
    height: 70px;
    padding: 0 10px 0 0;
    background-color: #1a73e8;
    user-select: none;
  }

  .brand {
    padding-left: 20px;
    margin: auto;
    font-size: 25px;
    letter-spacing: 0.3px;
    color: white;
  }

  .brand:hover {
    color: white;
    text-decoration: none;
  }

  .topbar .nav-link {
    padding-left: 8px;
    padding-right: 8px;
    color: white !important;
  }

  .topbar .link:hover {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 5px;
    cursor: pointer;
  }

  span {
    cursor: default;
  }

  @media (max-width: 900px) {
    .brand {
      padding-left: 90px;
    }
  }
</style>
