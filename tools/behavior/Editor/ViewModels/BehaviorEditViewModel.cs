using Editor.Framework;

namespace Editor.ViewModels
{
    class BehaviorEditViewModel : ViewModel
    {
        bool isReadOnly = false;

        public bool IsReadOnly { get { return isReadOnly; } set { SetProperty(ref isReadOnly, value); } }
    }
}
