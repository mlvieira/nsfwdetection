<script>
  import Router from "svelte-spa-router";
  import routes from "./router";
  import Navbar from "./components/Navbar.svelte";

  let showMenu = false;

  function toggleNavbar() {
    showMenu = !showMenu;
  }
</script>

<div class="container mx-auto p-4">
  <Navbar {toggleNavbar} {showMenu} />
  <Router {routes} />
</div>

<style>
</style>
