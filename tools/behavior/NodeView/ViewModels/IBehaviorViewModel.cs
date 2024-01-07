using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.ViewModels
{
    public interface IBehaviorViewModel
    {
        SimpleCommand AddItemCommand { get; }
        SimpleCommand RemoveItemCommand { get; }
        SimpleCommand ClearSelectedItemsCommand { get; }
        List<SelectableBehaviorItemViewModelBase> SelectedItems { get; }
        ObservableCollection<SelectableBehaviorItemViewModelBase> Items { get; }
    }
}
