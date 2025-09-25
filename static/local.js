function formatRelativeTime(rfc3339String) {
    const date = new Date(rfc3339String);
    const now = new Date();
    const diffMs = now - date;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    // Handle future dates
    if (diffMs < 0) {
        const futureDiffMs = Math.abs(diffMs);
        const futureDiffMins = Math.floor(futureDiffMs / 60000);
        const futureDiffHours = Math.floor(futureDiffMins / 60);
        const futureDiffDays = Math.floor(futureDiffHours / 24);

        if (futureDiffMins < 60) {
            return `in ${futureDiffMins} minute${futureDiffMins !== 1 ? 's' : ''}`;
        } else if (futureDiffHours < 24) {
            return `in ${futureDiffHours} hour${futureDiffHours !== 1 ? 's' : ''}`;
        } else {
            return `in ${futureDiffDays} day${futureDiffDays !== 1 ? 's' : ''}`;
        }
    }

    // Past dates
    if (diffSecs < 60) {
        return 'just now';
    } else if (diffMins < 60) {
        return diffMins === 1 ? '1 minute ago' : `${diffMins} minutes ago`;
    } else if (diffHours < 24) {
        return diffHours === 1 ? '1 hour ago' : `${diffHours} hours ago`;
    } else if (diffDays < 7) {
        return diffDays === 1 ? '1 day ago' : `${diffDays} days ago`;
    } else if (diffDays < 30) {
        const weeks = Math.floor(diffDays / 7);
        return weeks === 1 ? '1 week ago' : `${weeks} weeks ago`;
    } else if (diffDays < 365) {
        const months = Math.floor(diffDays / 30);
        return months === 1 ? '1 month ago' : `${months} months ago`;
    } else {
        const years = Math.floor(diffDays / 365);
        return years === 1 ? '1 year ago' : `${years} years ago`;
    }
}

// Replace timestamps on the page
function replaceTimestamps() {
    const timestampElements = document.querySelectorAll('[datetime], time');

    timestampElements.forEach(element => {
        const rfc3339String = element.getAttribute('datetime') || element.textContent.trim();

        // RFC3339 validation pattern
        if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?(Z|[+-]\d{2}:\d{2})$/.test(rfc3339String)) {
            const relativeTime = formatRelativeTime(rfc3339String);

            // Store original timestamp as title
            element.title = new Date(rfc3339String).toLocaleString();
            element.textContent = relativeTime;
            element.classList.add('relative-time');
        }
    });
}

// Auto-update every minute
function startAutoUpdate() {
    replaceTimestamps();
    setInterval(replaceTimestamps, 60000);
}

// Initialize
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', startAutoUpdate);
} else {
    startAutoUpdate();
}
