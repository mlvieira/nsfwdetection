<script>
    import { onMount } from "svelte";
    import { token } from "../stores/auth";
    import {
        closeWebSocket,
        initWebSocket,
        isWebSocketConnected,
    } from "../services/ws";
    import { push } from "svelte-spa-router";

    export let showMenu = false;
    export let toggleNavbar;

    function logout(event) {
        event.preventDefault();
        token.set("");
        closeWebSocket();
        push("/login");
    }

    onMount(() => {
        const unsubscribe = token.subscribe((jwtToken) => {
            if (jwtToken) {
                initWebSocket(jwtToken);
            } else {
                closeWebSocket();
            }
        });

        return () => {
            closeWebSocket();
        };
    });
</script>

<div>
    <nav
        class="container px-6 py-8 mx-auto md:flex md:justify-between md:items-center"
    >
        <div class="flex items-center justify-between w-full md:w-auto">
            <a
                class="text-xl font-bold text-gray-800 md:text-2xl hover:text-blue-400"
                href="/home"
                >Logo
            </a>

            <div class="flex items-center space-x-4 md:hidden">
                {#if $token}
                    <div
                        class={`h-3 w-3 rounded-full motion-safe:animate-pulse ${
                            isWebSocketConnected ? "bg-green-500" : "bg-red-500"
                        }`}
                    ></div>
                {/if}

                <button
                    type="button"
                    on:click={toggleNavbar}
                    class="text-gray-800 hover:text-gray-400 focus:outline-none focus:text-gray-400"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                        class="w-6 h-6"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
                        />
                    </svg>
                </button>
            </div>
        </div>

        <div
            class={`flex-col mt-8 space-y-4 md:flex md:space-y-0 md:flex-row md:items-center md:space-x-10 md:mt-0 ${
                showMenu ? "flex" : "hidden"
            }`}
        >
            {#if !$token}
                <a class="text-gray-800 hover:text-blue-400" href="/#">Login</a>
            {/if}

            {#if $token}
                <a class="text-gray-800 hover:text-blue-400" href="/#/label"
                    >Label</a
                >
                <a class="text-gray-800 hover:text-blue-400" href="/#/stats"
                    >Stats</a
                >
                <a
                    class="text-gray-800 hover:text-blue-400"
                    href="/"
                    on:click={logout}>Logout</a
                >

                <div
                    class={`flex items-center space-x-2 p-2 rounded-lg ${
                        isWebSocketConnected ? "bg-green-500" : "bg-red-500"
                    }`}
                >
                    <div
                        class="h-3 w-3 rounded-full motion-safe:animate-pulse bg-white"
                    ></div>

                    <span class="text-white font-semibold">
                        {isWebSocketConnected ? "Connected" : "Disconnected"}
                    </span>
                </div>
            {/if}
        </div>
    </nav>
</div>
