using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Media;

namespace Editor.Framework
{
    class PaneViewModel : ViewModel
    {
        #region 字段
        private string _title = null;
        private string _contentId = null;
        private bool _isSelected = false;
        private bool _isActive = false;
        #endregion

        #region 构造
        public PaneViewModel()
        {
        }
        #endregion

        #region 属性
        public string Title
        {
            get => _title;
            set => SetProperty(ref _title, value);
        }

        public ImageSource IconSource { get; protected set; }

        public string ContentId
        {
            get => _contentId;
            set => SetProperty(ref _contentId, value);
        }

        public bool IsSelected
        {
            get => _isSelected;
            set => SetProperty(ref _isSelected, value);
        }

        public bool IsActive
        {
            get => _isActive;
            set => SetProperty(ref _isActive, value);
        }
        #endregion
    }
}