window.goshtosoSyncDialog = (dialog, open) => {
  if (open && dialog.isConnected && !dialog.open) {
    dialog.showModal();
  }
  if (!open && dialog.open) {
    dialog.close();
  }
};
