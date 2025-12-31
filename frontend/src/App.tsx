import "./App.css";

export const App = () => {
  return (
    <div class="layout">
      <div class="layout__sidebar">
        <Sidebar />
      </div>
      <div class="layout__content px-4 py-3">right side</div>
    </div>
  );
};

const Sidebar = () => {
  return (
    <div class="px-4 py-3">
      <div class="flex gap-2 justify-between items-center">
        <div class="text-sm font-medium">Мои сервера</div>

        <button class="btn btn-xs btn-primary">Добавить</button>
      </div>
    </div>
  );
};
