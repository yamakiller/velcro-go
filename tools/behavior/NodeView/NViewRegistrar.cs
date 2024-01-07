using Splat;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeView
{
    public sealed class NViewRegistrar
    {
        private static readonly List<Tuple<Func<object>, Type>> PendingRegistrations = new List<Tuple<Func<object>, Type>>();
        private static Action<Func<object>, Type> m_registerAction;

        public static void AddRegistration(Func<object> factory, Type serviceType)
        {
            if (factory == null)
            {
                throw new ArgumentNullException(nameof(factory));
            }
            else if (serviceType == null)
            {
                throw new ArgumentNullException(nameof(serviceType));
            }

            if (m_registerAction == null)
            {
                PendingRegistrations.Add(Tuple.Create(factory, serviceType));
            }
            else
            {
                m_registerAction(factory, serviceType);
            }
        }

        public static void RegisterToLocator(Action<Func<object>, Type> newRegisterAction)
        {
            if (newRegisterAction == null)
            {
                throw new ArgumentNullException(nameof(newRegisterAction));
            }
            else if (m_registerAction != null)
            {
                throw new InvalidOperationException("A locator has already been set");
            }

            m_registerAction = newRegisterAction;
            foreach (var t in PendingRegistrations)
            {
                m_registerAction(t.Item1, t.Item2);
            }
            PendingRegistrations.Clear();
        }

        /// <summary>
        /// Register all NodeNetwork view/viewmodel pairs to Locator.CurrentMutable.
        /// </summary>
        public static void RegisterSplat()
        {
            RegisterToLocator((f, t) => Locator.CurrentMutable.Register(f, t));
        }
    }
}
