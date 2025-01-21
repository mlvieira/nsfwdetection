<script>
  import { onMount } from "svelte";
  import { token } from "../stores/auth";
  import { loginUser } from "../services/api";
  import { push } from "svelte-spa-router";
  import { showToast } from "../utils/toast";

  let username = "";
  let password = "";

  async function handleLogin() {
    try {
      const data = await loginUser(username, password);
      token.set(data.token);
      push("/label");
    } catch (err) {
      showToast(err.message || "Failed to login", "error");
    }
  }

  onMount(() => {
    token.subscribe((value) => {
      if (value) push("/label");
    });
  });
</script>

<div class="max-w-sm mx-auto mt-10 p-6 bg-white shadow rounded">
  <form on:submit|preventDefault={handleLogin} class="space-y-4">
    <div class="mb-4">
      <label class="block mb-1 font-semibold">Username</label>
      <input
        type="text"
        bind:value={username}
        class="border p-2 w-full rounded"
        required
      />
    </div>

    <div class="mb-4">
      <label class="block mb-1 font-semibold">Password</label>
      <input
        type="password"
        bind:value={password}
        class="border p-2 w-full rounded"
        required
        on:keydown={(e) => e.key === "Enter" && handleLogin()}
      />
    </div>

    <button
      type="submit"
      class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 w-full cursor-pointer"
      disabled={!username || !password}
    >
      Login
    </button>
  </form>
</div>
