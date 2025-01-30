import Login from './routes/Login.svelte';
import Label from './routes/Label.svelte';
import Stats from './routes/Stats.svelte';

const routes = {
    '/': Login,
    '/login': Login,
    '/label': Label,
    '/stats': Stats,
};

export default routes;
