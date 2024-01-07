using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.ViewModels
{
    public interface ISelectItems
    {
        SimpleCommand SelectItemCommand { get; }
    }

    public abstract class SelectableBehaviorItemViewModelBase : INotifyPropertyChangedBase, ISelectItems
    {
        private bool isSelected;

        public SelectableBehaviorItemViewModelBase(int id, IBehaviorViewModel parent)
        {
            this.Id = id;
            this.Parent = parent;
            Init();
        }

        public SelectableBehaviorItemViewModelBase()
        {
            Init();
        }

        public List<SelectableBehaviorItemViewModelBase> SelectedItems
        {
            get { return Parent.SelectedItems; }
        }

        public IBehaviorViewModel Parent { get; set; }
        public SimpleCommand SelectItemCommand { get; private set; }
        public int Id { get; set; }

        public bool IsSelected
        {
            get
            {
                return isSelected;
            }
            set
            {
                if (isSelected != value)
                {

                    isSelected = value;
                    NotifyChanged("IsSelected");
                }
            }
        }

        private void ExecuteSelectItemCommand(object param)
        {
            SelectItem((bool)param, !IsSelected);
        }

        private void SelectItem(bool newselect, bool select)
        {
            if (newselect)
            {
                foreach (var designerItemViewModelBase in Parent.SelectedItems.ToList())
                {
                    designerItemViewModelBase.isSelected = false;
                }
            }

            IsSelected = select;
        }


        private void Init()
        {
            SelectItemCommand = new SimpleCommand(ExecuteSelectItemCommand);
        }
    }
}
