function formatRelativeTime(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    // Handle future dates
    if (diffMs < 0) {
        return formatFutureTime(date, now);
    }

    // Past dates
    if (diffSecs < 60) {
        return 'just now';
    } else if (diffMins < 60) {
        return diffMins === 1 ? '1 minute ago' : `${diffMins} minutes ago`;
    } else if (diffHours < 24) {
        return diffHours === 1 ? '1 hour ago' : `${diffHours} hours ago`;
    } else if (diffDays === 0) {
        return `today at ${formatTime(date)}`;
    } else if (diffDays === 1) {
        return `yesterday at ${formatTime(date)}`;
    } else if (diffDays < 7) {
        return `${diffDays} days ago`;
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

function formatFutureTime(date, now) {
    const diffMs = date - now;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 60) {
        return `in ${diffMins} minute${diffMins !== 1 ? 's' : ''}`;
    } else if (diffHours < 24) {
        return `in ${diffHours} hour${diffHours !== 1 ? 's' : ''}`;
    } else if (diffDays < 7) {
        return `in ${diffDays} day${diffDays !== 1 ? 's' : ''}`;
    } else {
        return `on ${date.toLocaleDateString()}`;
    }
}

function formatTime(date) {
    return date.toLocaleTimeString('en-US', { 
        hour: 'numeric', 
        minute: '2-digit',
        hour12: true 
    });
}

// Function to replace all timestamp elements on the page
function replaceTimestamps() {
    // Select elements containing timestamps (adjust selector as needed)
    const timestampElements = document.querySelectorAll('[data-timestamp], time, .timestamp');

    timestampElements.forEach(element => {
        // Try to get timestamp from various sources
        let dateString = element.getAttribute('datetime') || 
                        element.getAttribute('data-timestamp') || 
                        element.textContent.trim();

        // Check if it looks like a timestamp
        if (isValidTimestamp(dateString)) {
            const relativeTime = formatRelativeTime(dateString);

            // Store original timestamp as title for hover
            element.title = new Date(dateString).toLocaleString();

            // Replace content with relative time
            element.textContent = relativeTime;

            // Add a class for styling if needed
            element.classList.add('relative-time');
        }
    });
}

function isValidTimestamp(str) {
    // Basic check for timestamp patterns
    const patterns = [
        /^\d{4}-\d{2}-\d{2}/, // ISO date format
        /^\d{1,2}\/\d{1,2}\/\d{4}/, // US date format
        /^\d{10,13}$/ // Unix timestamp
    ];

    return patterns.some(pattern => pattern.test(str)) && !isNaN(Date.parse(str));
}

// Auto-update timestamps every minute
function startAutoUpdate() {
    replaceTimestamps();
    setInterval(replaceTimestamps, 60000); // Update every minute
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', startAutoUpdate);
} else {
    startAutoUpdate();
}
