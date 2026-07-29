// Authored Alpine dark-mode store.
document.addEventListener('alpine:init', () => {
    const storageAllowed = () => !window.goshtosoStorageConsent || window.goshtosoStorageConsent.allowed();

    Alpine.store('darkMode', {
        // Initialize from localStorage or browser preference
        on: (() => {
            if (storageAllowed()) {
                const saved = localStorage.getItem('darkMode');
                if (saved !== null) {
                    return saved === 'true';
                }
            }
            // Default to browser preference
            return window.matchMedia('(prefers-color-scheme: dark)').matches;
        })(),

        init() {
            // Apply initial state to HTML element
            if (this.on) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        },

        toggle() {
            this.on = !this.on;
            if (storageAllowed()) {
                localStorage.setItem('darkMode', this.on);
            } else {
                localStorage.removeItem('darkMode');
            }

            if (this.on) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        }
    });

    // Initialize on load
    Alpine.store('darkMode').init();
});
