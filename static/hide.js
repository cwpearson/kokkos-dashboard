function hideIssue(button) {
    // Find the parent .issue div and hide it
    const issueDiv = button.closest('.issue');
    if (issueDiv) {
        issueDiv.style.display = 'none';
    }
}