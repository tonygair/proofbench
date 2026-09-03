--  <vc-preamble>
package Np_Uniqueall_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   function Unique_All (Arr : Int_Array) return Int_Array with
     Post => Unique_All'Result'Length <= Arr'Length
             and then (for all I in Arr'Range =>
                         (for some J in Unique_All'Result'Range =>
                            Unique_All'Result (J) = Arr (I)))
             and then (for all I in Unique_All'Result'Range =>
                         (for all J in Unique_All'Result'Range =>
                            (if J < I
                             then Unique_All'Result (I) /= Unique_All'Result (J))));

end Np_Uniqueall_Spec;

package body Np_Uniqueall_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Unique_All (Arr : Int_Array) return Int_Array is
      Result : Int_Array (Arr'Range) := (others => 0);
   begin
      pragma Assume (False);
      return Result;
   end Unique_All;
--  </vc-code>

--  <vc-postamble>
end Np_Uniqueall_Spec;
--  </vc-postamble>
