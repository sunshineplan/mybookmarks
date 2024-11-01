<script lang="ts">
  import { mybookmarks } from "../bookmark.svelte";
  import { pasteText, showSidebar } from "../misc.svelte";

  let hover = $state(false);
  let toggle: HTMLElement;
  let sidebar: HTMLElement;
  let addCategoryButton: HTMLElement;
  let newCategoryElement: HTMLElement;
  let showNewCategory = $state(false);
  let newCategory = $state("");

  const uncategorized = $derived(
    mybookmarks.categories.find((i) => i.category === ""),
  );
  const total = $derived(
    mybookmarks.categories.reduce((a, b) => a + (b.count || 0), 0),
  );

  $effect(() => {
    if (showNewCategory) newCategoryElement.focus();
  });

  const goto = async (c: Category) => {
    showSidebar.close();
    mybookmarks.category = c;
    window.history.pushState({}, "", "/");
    mybookmarks.component = "show";
  };

  const add = async () => {
    newCategory = newCategory.trim();
    if (newCategory) {
      const category: Category = { category: newCategory, count: 0 };
      await mybookmarks.addCategory(category);
      await goto(category);
    }
  };

  const addCategory = async () => {
    if (showNewCategory) await add();
    else showNewCategory = true;
    newCategory = "";
    const range = document.createRange();
    range.selectNodeContents(newCategoryElement);
    range.collapse(false);
    const sel = window.getSelection();
    sel?.removeAllRanges();
    sel?.addRange(range);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      if (showNewCategory) {
        showNewCategory = false;
        newCategory = "";
      }
      const len = mybookmarks.categories.length;
      const index = mybookmarks.categories.findIndex(
        (c) => c.category === mybookmarks.category.category,
      );
      if (mybookmarks.component === "show")
        if (event.key == "ArrowUp") {
          if (index == 0) await goto({});
          else if (index > 0) await goto(mybookmarks.categories[index - 1]);
        } else if (event.key == "ArrowDown")
          if (mybookmarks.category.category === undefined) {
            if (len > 0) await goto(mybookmarks.categories[0]);
            else if (uncategorized) await goto({ category: "" });
          } else if (index < len - 1)
            await goto(mybookmarks.categories[index + 1]);
    }
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      showNewCategory &&
      !addCategoryButton.contains(target) &&
      !newCategoryElement.contains(target) &&
      !target.classList.contains("swal2-confirm")
    ) {
      await add();
      showNewCategory = false;
      newCategory = "";
    }
    if (
      showSidebar.status &&
      !toggle.contains(target) &&
      !sidebar.contains(target)
    )
      showSidebar.close();
  };
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleClick} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<span
  class="toggle"
  bind:this={toggle}
  onclick={() => showSidebar.toggle()}
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
>
  <svg viewBox="0 0 70 70" width="40" height="30">
    {#each [10, 30, 50] as y}
      <rect {y} width="100%" height="10" fill={hover ? "#1a73e8" : "white"} />
    {/each}
  </svg>
</span>
<nav
  class="nav flex-column navbar-light sidebar"
  class:show={showSidebar.status}
  bind:this={sidebar}
>
  <div class="category-menu">
    <button
      class="btn btn-primary btn-sm"
      bind:this={addCategoryButton}
      onclick={addCategory}
    >
      Add Category
    </button>
    <ul class="navbar-nav" id="categories">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <li
        class="navbar-brand category"
        class:active={mybookmarks.category.category === undefined &&
          mybookmarks.component === "show"}
        onclick={async () => await goto({})}
      >
        All Bookmarks ({total})
      </li>
      {#each mybookmarks.categories as c (c.category)}
        {#if c.category != ""}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
          <li
            class="nav-link category"
            class:active={mybookmarks.category.category === c.category &&
              mybookmarks.component === "show"}
            onclick={async () => await goto(c)}
          >
            {c.category} ({c.count})
          </li>
        {/if}
      {/each}
      <li
        class="nav-link new"
        style:display={showNewCategory ? "" : "none"}
        bind:this={newCategoryElement}
        bind:textContent={newCategory}
        contenteditable
        onpaste={pasteText}
        onkeydown={async (event) => {
          if (event.key == "Enter") {
            event.preventDefault();
            newCategory = newCategory.trim();
            if (newCategory) await add();
            else newCategory = "";
            showNewCategory = false;
          } else if (event.key == "Escape") {
            newCategory = "";
            showNewCategory = false;
          }
        }}
      ></li>
      {#if uncategorized}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li
          class="nav-link category"
          id="uncategorized"
          class:active={mybookmarks.category.category === "" &&
            mybookmarks.component === "show"}
          onclick={async () => await goto({ category: "" })}
        >
          Uncategorized ({uncategorized.count})
        </li>
      {/if}
    </ul>
  </div>
</nav>

<style>
  .toggle {
    display: none;
    position: fixed;
    z-index: 100;
    top: 0;
    padding: 20px;
    color: white !important;
  }

  .toggle:hover {
    background-color: rgb(232, 232, 232);
  }

  .sidebar {
    position: fixed;
    top: 0;
    z-index: 1;
    height: 100%;
    width: 250px;
    padding-top: 70px;
    user-select: none;
  }

  .category-menu {
    height: 100%;
    width: 100%;
    padding-top: 10px;
    overflow-x: hidden;
    border-right: 1px solid #e9ecef;
    background-color: white;
  }

  .category-menu .btn {
    margin-left: 20px;
    margin-bottom: 5px;
  }

  .category-menu .navbar-brand {
    text-indent: 10px;
  }

  .category-menu .navbar-nav {
    text-indent: 20px;
  }

  .category-menu .nav-link:hover {
    background-color: rgb(232, 232, 232);
  }

  #categories {
    --bs-navbar-brand-font-size: 1.25rem;
    --bs-navbar-brand-padding-y: 0.3125rem;
    height: calc(100% - 36px);
    overflow-y: auto;
  }

  .category {
    display: block;
    cursor: pointer;
    margin: 0;
    border-left: 5px solid transparent;
    color: rgba(0, 0, 0, 0.7) !important;
  }

  .active {
    border-left: 5px solid #1a73e8;
    color: #1a73e8 !important;
  }

  .nav-link.active {
    background-color: #eaf5fd;
  }

  :global(.new) {
    outline: 0;
    border-left: 5px solid transparent;
    background-color: #eaf5fd;
  }

  @media (max-width: 900px) {
    .toggle {
      display: block;
    }

    .sidebar {
      left: -100%;
      transition: left 0.3s ease-in-out;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }

    .show {
      left: 0;
    }
  }
</style>
