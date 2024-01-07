using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.Helpers
{
    public sealed class WeakINotifyEventHandler
    {
        private readonly WeakReference m_targetReference;
        private readonly MethodInfo m_method;

        public WeakINotifyEventHandler(PropertyChangedEventHandler callback)
        {
            m_method = callback.Method;
            m_targetReference = new WeakReference(callback.Target, true);
        }

        //[DebuggerNonUserCode]
        public void Handler(object sender, PropertyChangedEventArgs e)
        {
            var target = m_targetReference.Target;
            if (target != null)
            {
                var callback = (Action<object, PropertyChangedEventArgs>)Delegate.CreateDelegate(typeof(Action<object, PropertyChangedEventArgs>), target, _method, true);
                if (callback != null)
                {
                    callback(sender, e);
                }
            }
        }
    }
}
